// @vitest-environment node
import { ESLint } from "eslint";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

// Страховка правила «нет текста в JSX»: i18next/no-literal-string действительно ловит
// текст между тегами в src/**/*.tsx. Если правило удалят из eslint.config.js или
// обновление плагина тихо его выключит, упадёт этот тест, а не ревью.
const cwd = fileURLToPath(new URL("../..", import.meta.url));
const eslint = new ESLint({ cwd });
const RULE = "i18next/no-literal-string";

const literalErrors = async (code: string, file = "src/pages/Probe.tsx") => {
  const [result] = await eslint.lintText(code, { filePath: `${cwd}/${file}` });
  return (result?.messages ?? []).filter((m) => m.ruleId === RULE);
};

describe("правило no-literal-string", () => {
  it("включено для компонентов", async () => {
    const config = (await eslint.calculateConfigForFile(`${cwd}/src/app/Shell.tsx`)) as {
      rules?: Record<string, unknown>;
    };
    expect(config.rules?.[RULE]).toEqual([2, { mode: "jsx-text-only" }]);
  });

  it("ловит текст между тегами", async () => {
    const errors = await literalErrors(
      "export const Probe = () => (\n  <div>\n    <p>Привет мир</p>\n  </div>\n);\n",
    );
    expect(errors).toHaveLength(1);
  });

  it("не ругается на t() и строковые пропсы", async () => {
    const errors = await literalErrors(
      'import { useI18n } from "../i18n";\n' +
        "export const Probe = () => {\n" +
        "  const { t } = useI18n();\n" +
        '  return <p className="x">{t("settings.title")}</p>;\n' +
        "};\n",
    );
    expect(errors).toHaveLength(0);
  });
});
