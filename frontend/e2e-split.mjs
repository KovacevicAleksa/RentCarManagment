import { chromium } from "playwright";
import { readFileSync } from "node:fs";

const USER_APP = "http://localhost:5273";
const ADMIN_APP = "http://localhost:5274";
const SHOTS = "/tmp/rc_shots";
const regUser = readFileSync(new URL("./.e2e_split_user", import.meta.url), "utf8").trim();

let pass = 0, fail = 0;
const ok = (n) => { pass++; console.log(`  \x1b[32mPASS\x1b[0m ${n}`); };
const bad = (n, e) => { fail++; console.log(`  \x1b[31mFAIL\x1b[0m ${n}${e ? " — " + e : ""}`); };

const login = async (page, email, pw) => {
  await page.getByPlaceholder("ime@primer.com").fill(email);
  await page.getByPlaceholder("••••••••").fill(pw);
  await page.getByRole("button", { name: "Prijavi se" }).click();
};

const browser = await chromium.launch();

// ───────────────────────── USER APP (5273) ─────────────────────────
{
  const page = await browser.newPage();
  try {
    console.log("\n══ USER APP (:5273) ══");
    await page.goto(USER_APP, { waitUntil: "networkidle" });
    // even an admin logging into the USER app gets no admin capability
    await login(page, "admin@rentcar.com", "admin12345");
    await page.getByText("Kontrolna tabla voznog parka").waitFor({ timeout: 10000 });
    ok("user app: login → dashboard");

    await page.locator("header button").last().click();
    await page.getByRole("button", { name: "Odjavi se" }).waitFor({ timeout: 4000 });
    const hasAdminBtn = await page.getByRole("button", { name: "Administracija" }).count();
    const hasSettings = await page.getByRole("button", { name: "Podešavanja" }).count();
    hasAdminBtn === 0 ? ok("user app: NEMA 'Administracija' u meniju") : bad("user app: admin dugme i dalje postoji");
    hasSettings === 1 ? ok("user app: 'Podešavanja' postoji") : bad("user app: settings dugme");

    // Settings → Sistem: no 'Otvori administraciju' even for an admin account
    await page.getByRole("button", { name: "Podešavanja" }).click();
    await page.getByRole("heading", { name: "Podešavanja" }).waitFor({ timeout: 5000 });
    await page.getByRole("button", { name: "Sistem" }).click();
    await page.getByRole("heading", { name: "Sistem i sesija" }).waitFor();
    const hasOpenAdmin = await page.getByRole("button", { name: "Otvori administraciju" }).count();
    hasOpenAdmin === 0 ? ok("user app: Settings NEMA 'Otvori administraciju'") : bad("user app: open-admin link postoji");
    await page.screenshot({ path: `${SHOTS}/split-user-app.png`, fullPage: true });
  } catch (e) {
    bad("user app: neočekivana greška", e.message);
    await page.screenshot({ path: `${SHOTS}/split-user-error.png` }).catch(() => {});
  } finally {
    await page.close();
  }
}

// ─────────────────────── ADMIN APP (5274) ───────────────────────
{
  const page = await browser.newPage();
  try {
    console.log("\n══ ADMIN APP (:5274) ══");
    await page.goto(ADMIN_APP, { waitUntil: "networkidle" });
    (await page.getByRole("heading", { name: "RentCar Admin" }).count()) > 0
      ? ok("admin app: prikazan admin login ('RentCar Admin')")
      : bad("admin app: admin login branding");
    await page.screenshot({ path: `${SHOTS}/split-admin-login.png` });

    // admin login → AdminPanel directly (no dashboard)
    await login(page, "admin@rentcar.com", "admin12345");
    await page.getByRole("heading", { name: "Administracija" }).waitFor({ timeout: 10000 });
    await page.getByText("Novi nalog").waitFor();
    ok("admin app: admin login → AdminPanel direktno");
    (await page.getByText("Kontrolna tabla voznog parka").count()) === 0
      ? ok("admin app: NEMA fleet dashboard-a")
      : bad("admin app: dashboard se pojavio");
    (await page.getByRole("button", { name: "Odjavi se" }).count()) > 0
      ? ok("admin app: header ima 'Odjavi se' (standalone)")
      : bad("admin app: logout u headeru");

    // create an account from the admin app
    const newEmail = `fromadmin_${Date.now()}@rentcar.com`;
    await page.getByPlaceholder("ime@primer.com").fill(newEmail);
    await page.getByPlaceholder("Najmanje 6 karaktera").fill("password123");
    await page.getByRole("button", { name: "Kreiraj nalog" }).click();
    await page.getByText(new RegExp(`Nalog ${newEmail} je kreiran`)).waitFor({ timeout: 8000 });
    ok("admin app: kreiranje naloga radi");
    await page.screenshot({ path: `${SHOTS}/split-admin-panel.png`, fullPage: true });

    // logout → back to admin login
    await page.getByRole("button", { name: "Odjavi se" }).click();
    await page.getByRole("heading", { name: "RentCar Admin" }).waitFor({ timeout: 6000 });
    ok("admin app: odjava → admin login");

    // regular user → access denied
    await login(page, regUser, "password123");
    await page.getByRole("heading", { name: "Pristup odbijen" }).waitFor({ timeout: 8000 });
    ok("admin app: običan korisnik → 'Pristup odbijen' (RBAC)");
    (await page.getByText("Novi nalog").count()) === 0
      ? ok("admin app: običan korisnik NE vidi admin formu")
      : bad("admin app: user video admin formu!");
    await page.screenshot({ path: `${SHOTS}/split-admin-denied.png` });
  } catch (e) {
    bad("admin app: neočekivana greška", e.message);
    await page.screenshot({ path: `${SHOTS}/split-admin-error.png` }).catch(() => {});
  } finally {
    await page.close();
  }
}

await browser.close();
console.log(`\n REZULTAT: \x1b[32m${pass} prošlo\x1b[0m, \x1b[31m${fail} palo\x1b[0m`);
process.exit(fail);
