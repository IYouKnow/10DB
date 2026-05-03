import { describe, expect, it } from "vitest";
import { hasTables } from "./validators";

describe("hasTables", () => {
  it("returns false for an empty blueprint", () => {
    expect(
      hasTables({
        version: 1,
        projectId: "proj_1",
        tables: []
      })
    ).toBe(false);
  });

  it("returns true when the blueprint has at least one table", () => {
    expect(
      hasTables({
        version: 1,
        projectId: "proj_1",
        tables: [
          {
            id: "tbl_users",
            name: "users",
            position: { x: 0, y: 0 },
            columns: [],
            foreignKeys: []
          }
        ]
      })
    ).toBe(true);
  });
});
