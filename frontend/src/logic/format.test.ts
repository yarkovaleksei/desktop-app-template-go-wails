import { describe, expect, it } from "vitest";
import { formatBytes, minutesSince, progressPercent } from "./format";

describe("format", () => {
  it("formatBytes — мегабайты по языку", () => {
    expect(formatBytes("en", 48_200_000)).toBe("48.2 MB");
    expect(formatBytes("ru", 48_200_000)).toMatch(/^48,2\s*МБ$/);
  });

  it("progressPercent ограничен 0–100", () => {
    expect(progressPercent(50, 200)).toBe(25);
    expect(progressPercent(300, 200)).toBe(100);
    expect(progressPercent(10, 0)).toBe(0);
    expect(progressPercent(undefined, undefined)).toBe(0);
  });

  it("minutesSince — полные минуты", () => {
    expect(minutesSince(0, 59_999)).toBe(0);
    expect(minutesSince(0, 125_000)).toBe(2);
    expect(minutesSince(10, 0)).toBe(0);
  });
});
