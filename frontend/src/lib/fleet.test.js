import { describe, it, expect } from "vitest";
import { filterCars } from "./fleet";

const cars = {
  CAR003: { car_id: "CAR003" },
  CAR001: { car_id: "CAR001" },
  CAR012: { car_id: "CAR012" },
};

describe("filterCars", () => {
  it("returns all cars sorted by id when query is empty", () => {
    expect(filterCars(cars, "").map((c) => c.car_id)).toEqual([
      "CAR001",
      "CAR003",
      "CAR012",
    ]);
  });

  it("filters by a case-insensitive substring of the id", () => {
    expect(filterCars(cars, "car0").map((c) => c.car_id)).toEqual([
      "CAR001",
      "CAR003",
      "CAR012",
    ]);
    expect(filterCars(cars, "3").map((c) => c.car_id)).toEqual(["CAR003"]);
    expect(filterCars(cars, "CAR012").map((c) => c.car_id)).toEqual(["CAR012"]);
  });

  it("returns an empty array when nothing matches", () => {
    expect(filterCars(cars, "xyz")).toEqual([]);
  });

  it("tolerates a missing cars object", () => {
    expect(filterCars(undefined, "car")).toEqual([]);
    expect(filterCars(null, "")).toEqual([]);
  });
});
