import { chromium } from "playwright";
import { readFileSync } from "node:fs";

const BASE = "http://localhost:5273";
const SHOTS = "/tmp/rc_shots";
const target = readFileSync(new URL("./.e2e_reset_user", import.meta.url), "utf8").trim();

let pass = 0, fail = 0;
const ok = (n) => { pass++; console.log(`  \x1b[32mPASS\x1b[0m ${n}`); };
const bad = (n, e) => { fail++; console.log(`  \x1b[31mFAIL\x1b[0m ${n}${e ? " — " + e : ""}`); };

const b = await chromium.launch();
const page = await b.newPage();
page.on("dialog", (d) => d.accept()); // auto-accept the confirm()
try {
  // login as admin
  await page.goto(BASE, { waitUntil: "networkidle" });
  await page.getByPlaceholder("ime@primer.com").fill("admin@rentcar.com");
  await page.getByPlaceholder("••••••••").fill("admin12345");
  await page.getByRole("button", { name: "Prijavi se" }).click();
  await page.getByText("Kontrolna tabla voznog parka").waitFor({ timeout: 10000 });
  ok("admin ulogovan");

  // open admin panel
  await page.locator("header button").last().click();
  await page.getByRole("button", { name: "Administracija" }).click();
  await page.getByRole("heading", { name: "Administracija" }).waitFor({ timeout: 5000 });
  ok("admin panel otvoren");

  // find the target user's row and click Reset
  const row = page.locator("li").filter({ hasText: target });
  await row.first().waitFor({ timeout: 5000 });
  await row.first().getByRole("button", { name: "Reset" }).click();
  ok(`kliknut Reset za ${target}`);

  // capture the temporary password
  const tempEl = page.getByTestId("temp-password");
  await tempEl.waitFor({ timeout: 8000 });
  const temp = (await tempEl.innerText()).trim();
  await page.screenshot({ path: `${SHOTS}/r1-temp-password.png`, fullPage: true });
  temp.length === 12 ? ok(`privremena lozinka prikazana (${temp.length} char)`) : bad(`temp length ${temp.length}`);

  // back to dashboard, then logout
  await page.getByRole("button", { name: "Nazad" }).click();
  await page.getByText("Kontrolna tabla voznog parka").waitFor({ timeout: 8000 });
  await page.locator("header button").last().click();
  await page.getByRole("button", { name: "Odjavi se" }).click();
  await page.getByRole("button", { name: "Prijavi se" }).waitFor({ timeout: 6000 });
  ok("admin odjavljen");

  // login as the target user with the temporary password
  await page.getByPlaceholder("ime@primer.com").fill(target);
  await page.getByPlaceholder("••••••••").fill(temp);
  await page.getByRole("button", { name: "Prijavi se" }).click();
  await page.getByText("Kontrolna tabla voznog parka").waitFor({ timeout: 10000 });
  ok("korisnik se ulogovao PRIVREMENOM lozinkom (end-to-end)");
} catch (e) {
  bad("neočekivana greška", e.message);
  await page.screenshot({ path: `${SHOTS}/r-error.png` }).catch(() => {});
} finally {
  await b.close();
}
console.log(`\n REZULTAT: \x1b[32m${pass} prošlo\x1b[0m, \x1b[31m${fail} palo\x1b[0m`);
process.exit(fail);
