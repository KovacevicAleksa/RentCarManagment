import { chromium } from "playwright";
import { readFileSync } from "node:fs";

const BASE = "http://localhost:5273";
const SHOTS = "/tmp/rc_shots";
const email = readFileSync(new URL("./.e2e_user", import.meta.url), "utf8").trim();

let pass = 0, fail = 0;
const ok = (n) => { pass++; console.log(`  \x1b[32mPASS\x1b[0m ${n}`); };
const bad = (n, e) => { fail++; console.log(`  \x1b[31mFAIL\x1b[0m ${n}${e ? " — " + e : ""}`); };

const b = await chromium.launch();
const page = await b.newPage();
try {
  // login as regular user
  await page.goto(BASE, { waitUntil: "networkidle" });
  await page.getByPlaceholder("ime@primer.com").fill(email);
  await page.getByPlaceholder("••••••••").fill("password123");
  await page.getByRole("button", { name: "Prijavi se" }).click();
  await page.getByText("Kontrolna tabla voznog parka").waitFor({ timeout: 10000 });
  ok("login običnog korisnika → dashboard");

  // open settings
  await page.locator("header button").last().click();
  await page.getByRole("button", { name: "Podešavanja" }).click();
  await page.getByRole("heading", { name: "Podešavanja" }).waitFor({ timeout: 5000 });
  ok("Settings stranica otvorena");

  // profile section
  (await page.getByText(email, { exact: true }).count()) > 0
    ? ok("Profil prikazuje email korisnika")
    : bad("Profil prikazuje email");
  await page.screenshot({ path: `${SHOTS}/s1-profile.png`, fullPage: true });

  // security: change password
  await page.getByRole("button", { name: "Bezbednost" }).click();
  await page.getByRole("heading", { name: "Promena lozinke" }).waitFor();
  const pw = page.locator('input[type="password"]');
  await pw.nth(0).fill("password123"); // current
  await pw.nth(1).fill("brandnew123"); // new
  await pw.nth(2).fill("brandnew123"); // confirm
  await page.getByRole("button", { name: "Promeni lozinku" }).click();
  await page.getByText("Lozinka je uspešno promenjena").waitFor({ timeout: 8000 });
  await page.screenshot({ path: `${SHOTS}/s2-password.png`, fullPage: true });
  ok("promena lozinke preko UI uspela");

  // security: password mismatch validation
  await pw.nth(0).fill("brandnew123");
  await pw.nth(1).fill("abcdef1");
  await pw.nth(2).fill("different1");
  await page.getByRole("button", { name: "Promeni lozinku" }).click();
  await page.getByText(/ne poklapaju/).waitFor({ timeout: 4000 });
  ok("validacija: lozinke se ne poklapaju");

  // display: change a threshold and save
  await page.getByRole("button", { name: "Prikaz" }).click();
  await page.getByRole("heading", { name: "Pragovi upozorenja" }).waitFor();
  const engineWarn = page.locator('label:has-text("Motor — upozorenje") + input');
  await engineWarn.fill("70");
  await page.getByRole("button", { name: "Sačuvaj" }).click();
  await page.getByText("Podešavanja prikaza su sačuvana").waitFor({ timeout: 5000 });
  await page.screenshot({ path: `${SHOTS}/s3-thresholds.png`, fullPage: true });
  ok("pragovi prikaza sačuvani");

  // invalid threshold (warn >= critical)
  await engineWarn.fill("250");
  await page.getByRole("button", { name: "Sačuvaj" }).click();
  await page.getByText(/između 0 i 200/).waitFor({ timeout: 4000 });
  ok("validacija: prag van opsega odbijen");

  // persistence: reset then reopen to confirm default restored
  await page.getByRole("button", { name: "Podrazumevano" }).click();
  await page.getByText("Vraćeno na podrazumevane").waitFor({ timeout: 4000 });
  const val = await engineWarn.inputValue();
  val === "90" ? ok("reset vraća podrazumevani prag (90)") : bad(`reset (viđeno ${val})`);

  // system section
  await page.getByRole("button", { name: "Sistem" }).click();
  await page.getByRole("heading", { name: "Sistem i sesija" }).waitFor();
  (await page.getByText("1.0.0").count()) > 0 ? ok("Sistem prikazuje verziju") : bad("verzija");
  (await page.getByRole("button", { name: "Otvori administraciju" }).count()) === 0
    ? ok("običan korisnik NE vidi 'Otvori administraciju' (RBAC)")
    : bad("RBAC: admin dugme skriveno za usera");
  await page.screenshot({ path: `${SHOTS}/s4-system.png`, fullPage: true });

  // logout from settings
  await page.getByRole("button", { name: "Odjavi se" }).click();
  await page.getByRole("button", { name: "Prijavi se" }).waitFor({ timeout: 6000 });
  ok("odjava iz Settings → login ekran");

  // verify new password works
  await page.getByPlaceholder("ime@primer.com").fill(email);
  await page.getByPlaceholder("••••••••").fill("brandnew123");
  await page.getByRole("button", { name: "Prijavi se" }).click();
  await page.getByText("Kontrolna tabla voznog parka").waitFor({ timeout: 10000 });
  ok("re-login sa NOVOM lozinkom uspeo (end-to-end)");
} catch (e) {
  bad("neočekivana greška", e.message);
  await page.screenshot({ path: `${SHOTS}/s-error.png` }).catch(() => {});
} finally {
  await b.close();
}
console.log(`\n REZULTAT: \x1b[32m${pass} prošlo\x1b[0m, \x1b[31m${fail} palo\x1b[0m`);
process.exit(fail);
