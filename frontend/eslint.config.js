import js from "@eslint/js";
import i18next from "eslint-plugin-i18next";
import reactHooks from "eslint-plugin-react-hooks";
import { defineConfig } from "eslint/config";
import globals from "globals";
import tseslint from "typescript-eslint";

export default defineConfig(
  { ignores: ["dist", "bindings", "node_modules", ".yarn", "coverage", "playwright-report", "test-results"] },
  {
    files: ["**/*.{js,ts,tsx}"],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: { "react-hooks": reactHooks },
    rules: {
      ...reactHooks.configs.recommended.rules,
    },
  },
  {
    // Никакого текста интерфейса прямо в JSX — только t() и словари i18n/ru.ts, i18n/en.ts.
    // Режим jsx-text-only ловит текст между тегами; строки в пропсах (variant="h6")
    // разрешены, иначе правило тонет в ложных срабатываниях. Тесты пишут текст в JSX свободно.
    // Что правило включено и работает, проверяет src/test/i18n-lint.test.ts.
    files: ["src/**/*.tsx"],
    ignores: ["src/**/*.test.tsx", "src/test/**"],
    plugins: { i18next },
    rules: {
      "i18next/no-literal-string": ["error", { mode: "jsx-text-only" }],
    },
  },
);
