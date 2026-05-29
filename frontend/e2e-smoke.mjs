import { chromium } from "playwright";
import { mkdirSync } from "node:fs";

const BASE = "http://localhost:5273";
const SHOTS = "/tmp/rc_shots";
mkdirSync(SHOTS, { recursive: true });

let pass = 0,
  fail = 0;
const ok = (n) => {
  pass++;
  console.log(`  \x1b[32mPASS\x1b[0m ${n}`);
};
const bad = (n, e) => {
  fail++;
  console.log(`  \x1b[31mFAIL\x1b[0m ${n}${e ? " — " + e : ""}`);
};

const uniqueEmail = `pw_${Date.now()}@rentcar.com`;

const browser = await chromium.launch();
const page = await browser.newPage();
const consoleErrors = [];
page.on("console", (m) => m.type() === "error" && consoleErrors.push(m.text()));
page.on("pageerror", (e) => consoleErrors.push(String(e)));

try {
  // 1. Load login page
  await page.goto(BASE, { waitUntil: "networkidle" });
  await page.screenshot({ path: `${SHOTS}/01-login.png` });
  (await page.getByRole("heading", { name: "RentCar" }).isVisible())
    ? ok("login stranica se učitala (RentCar logo)")
    : bad("login stranica se učitala");

  // 2. Login as admin
  await page.getByPlaceholder("ime@primer.com").fill("admin@rentcar.com");
  await page.getByPlaceholder("••••••••").fill("admin12345");
  await page.getByRole("button", { name: "Prijavi se" }).click();

  // 3. Dashboard appears
  await page.getByText("Kontrolna tabla voznog parka").waitFor({ timeout: 10000 });
  ok("login kao admin → dashboard prikazan");

  // 4. Live cars arrive via WebSocket
  await page.getByText(/CAR00\d/).first().waitFor({ timeout: 12000 });
  const carCount = await page.locator("h3").filter({ hasText: /^CAR00\d$/ }).count();
  await page.screenshot({ path: `${SHOTS}/02-dashboard.png`, fullPage: true });
  carCount > 0
    ? ok(`dashboard prikazuje žива vozila preko WS (${carCount} kartica)`)
    : bad("dashboard prikazuje vozila");
  carCount === 5
    ? ok("tačno 5 vozila prikazano")
    : bad(`tačno 5 vozila prikazano (viđeno ${carCount})`);

  // 5. Open user menu and find admin button
  await page.locator("header button").last().click();
  const adminBtn = page.getByRole("button", { name: "Administracija" });
  await adminBtn.waitFor({ timeout: 5000 });
  ok("dugme 'Administracija' vidljivo adminu");

  // 6. Open admin panel
  await adminBtn.click();
  await page.getByRole("heading", { name: "Administracija" }).waitFor({ timeout: 5000 });
  await page.getByText("Novi nalog").waitFor();
  await page.screenshot({ path: `${SHOTS}/03-admin-panel.png`, fullPage: true });
  ok("admin panel otvoren (forma + lista naloga)");

  // 7. List already shows existing users (incl. admin)
  const adminInList = await page.getByText("admin@rentcar.com").count();
  adminInList > 0
    ? ok("lista naloga prikazuje postojeće korisnike")
    : bad("lista naloga prikazuje korisnike");

  // 8. Create a new account
  await page.getByPlaceholder("ime@primer.com").fill(uniqueEmail);
  await page.getByPlaceholder("Najmanje 6 karaktera").fill("password123");
  await page.locator("select").selectOption("admin");
  await page.getByRole("button", { name: "Kreiraj nalog" }).click();

  // 9. Success message
  await page.getByText(new RegExp(`Nalog ${uniqueEmail} je kreiran`)).waitFor({ timeout: 8000 });
  ok(`nalog kreiran preko UI: ${uniqueEmail}`);

  // 10. New account appears in the list (exact match → the list span, not the toast)
  await page.getByText(uniqueEmail, { exact: true }).waitFor({ timeout: 5000 });
  await page.screenshot({ path: `${SHOTS}/04-after-create.png`, fullPage: true });
  ok("novi nalog se pojavio u listi");

  // 11. Validation: invalid email blocked client-side
  await page.getByPlaceholder("ime@primer.com").fill("not-an-email");
  await page.getByPlaceholder("Najmanje 6 karaktera").fill("password123");
  await page.getByRole("button", { name: "Kreiraj nalog" }).click();
  await page.getByText(/Unesite ispravnu email/).waitFor({ timeout: 4000 });
  ok("klijentska validacija blokira nevalidan email");

  // 12. Logout flow → back to login
  await page.getByRole("button", { name: "Nazad" }).click();
  await page.locator("header button").last().click();
  await page.getByRole("button", { name: "Odjavi se" }).click();
  await page.getByRole("button", { name: "Prijavi se" }).waitFor({ timeout: 6000 });
  await page.screenshot({ path: `${SHOTS}/05-after-logout.png` });
  ok("odjava → vraćen na login ekran");
} catch (e) {
  bad("neočekivana greška u toku", e.message);
  await page.screenshot({ path: `${SHOTS}/error.png` }).catch(() => {});
} finally {
  await browser.close();
}

console.log("\n────────────────────────────────────────────");
console.log(` Console/page greške u browseru: ${consoleErrors.length}`);
consoleErrors.slice(0, 10).forEach((e) => console.log("   • " + e));
console.log(`\n REZULTAT: \x1b[32m${pass} prošlo\x1b[0m, \x1b[31m${fail} palo\x1b[0m`);
console.log(` Screenshot-ovi: ${SHOTS}/`);
process.exit(fail);
