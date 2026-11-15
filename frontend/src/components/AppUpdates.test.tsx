import { act, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { UpdateInstall, UpdateState, type UpdateStatus } from "../api";
import { statusOf } from "../test/fakeSource";
import { renderApp, ruSettings } from "../test/render";

const MB = 1_000_000;
const available = { latest: "v1.1.0", package: "desktop-app_1.1.0_amd64.deb", size: 48.2 * MB };

const open = (update: UpdateStatus) => renderApp({ settings: ruSettings, update }, "/settings");
const button = (name: string | RegExp) => screen.getByRole("button", { name });

describe("AppUpdates", () => {
  it("unsupported: обновление вручную, кнопки проверки нет", async () => {
    await open(statusOf({ revision: 1, state: UpdateState.StateUnsupported, format: "" as never }));
    expect(screen.getByText("Эта сборка не из пакета и обновляется вручную.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Проверить обновления" })).toBeNull();
  });

  it("idle: проверка вызывает источник", async () => {
    const { source } = await open(statusOf({ revision: 1 }));
    await userEvent.click(button("Проверить обновления"));
    expect(source.checkUpdate).toHaveBeenCalledOnce();
    expect(button("Проверка…")).toBeDisabled();
  });

  it("current: установлена последняя версия", async () => {
    await open(statusOf({ revision: 1, state: UpdateState.StateCurrent, latest: "v1.0.0" }));
    expect(screen.getByText("Установлена последняя версия v1.0.0")).toBeInTheDocument();
    expect(button("Проверить обновления")).toBeEnabled();
  });

  it("available: версия, размер, скачивание и заметки", async () => {
    const { source } = await open(
      statusOf({ revision: 1, state: UpdateState.StateAvailable, ...available, notesUrl: "https://n" }),
    );
    expect(screen.getByText(/Доступна версия v1\.1\.0, размер 48,2\s*МБ/)).toBeInTheDocument();
    await userEvent.click(button("Что нового"));
    expect(source.openURL).toHaveBeenCalledWith("https://n");
    // Ответ fakeSource на скачивание — статус downloading без заметок: кнопка пропадает.
    await userEvent.click(button("Скачать"));
    expect(source.downloadUpdate).toHaveBeenCalledOnce();
    expect(screen.getByRole("progressbar")).toBeInTheDocument();
  });

  it("downloading: прогресс по событию и отмена", async () => {
    const { source, emitUpdate } = await open(
      statusOf({ revision: 1, state: UpdateState.StateDownloading, ...available, received: 0 }),
    );
    act(() =>
      emitUpdate(
        statusOf({ revision: 2, state: UpdateState.StateDownloading, ...available, received: 24.1 * MB }),
      ),
    );
    expect(screen.getByRole("progressbar")).toHaveAttribute("aria-valuenow", "50");
    expect(screen.getByText(/Скачано 24,1\s*МБ из 48,2\s*МБ/)).toBeInTheDocument();
    await userEvent.click(button("Отменить"));
    expect(source.cancelUpdate).toHaveBeenCalledOnce();
  });

  it("ready с командой: текст, копирование и папка с пакетом", async () => {
    const command = "sudo apt install /home/u/.cache/desktop-app/updates/desktop-app_1.1.0_amd64.deb";
    const { source } = await open(
      statusOf({
        revision: 1,
        state: UpdateState.StateReady,
        ...available,
        install: UpdateInstall.InstallCommand,
        command,
      }),
    );
    expect(
      screen.getByText(/Пакет v1\.1\.0 скачан и сверен\. Установите его командой ниже/),
    ).toBeInTheDocument();
    expect(screen.getByText(command)).toBeInTheDocument();
    await userEvent.click(button("Скопировать команду"));
    expect(source.copyText).toHaveBeenCalledWith(command);
    expect(await screen.findByText("Команда скопирована")).toBeInTheDocument();
    await userEvent.click(button("Открыть папку с пакетом"));
    expect(source.installUpdate).toHaveBeenCalledOnce();
  });

  it.each([
    [UpdateInstall.InstallOpen, "Открыть образ диска", /перетащите приложение в «Программы»/],
    [UpdateInstall.InstallRun, "Запустить установщик", /Установщик заменит приложение/],
    [UpdateInstall.InstallReplace, "Установить и перезапустить", /Файл AppImage заменится новым/],
    [UpdateInstall.InstallReveal, "Показать файл", /замените его файл скачанным/],
  ])("ready (%s): своя подсказка и кнопка", async (install, label, text) => {
    await open(statusOf({ revision: 1, state: UpdateState.StateReady, ...available, install }));
    expect(screen.getByText(text)).toBeInTheDocument();
    expect(button(label)).toBeEnabled();
    expect(screen.queryByRole("button", { name: "Скопировать команду" })).toBeNull();
  });

  it("failed: текст ошибки из словаря по коду, повтор скачивания", async () => {
    const { source } = await open(
      statusOf({
        revision: 1,
        state: UpdateState.StateFailed,
        ...available,
        error: { code: "update.network" },
      }),
    );
    expect(
      screen.getByText("Не удалось связаться с GitHub — проверьте подключение к сети"),
    ).toBeInTheDocument();
    await userEvent.click(button("Повторить"));
    expect(source.downloadUpdate).toHaveBeenCalledOnce();
  });

  it("failed с параметрами: подстановка в текст", async () => {
    await open(
      statusOf({
        revision: 1,
        state: UpdateState.StateFailed,
        error: { code: "update.no-package", params: { version: "v1.1.0", package: "x.rpm" } },
      }),
    );
    expect(
      screen.getByText("В выпуске v1.1.0 нет пакета x.rpm — обновите приложение вручную"),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Повторить" })).toBeNull();
  });
});
