import { chromium } from "playwright";

const ADMIN_APP = "http://localhost:5274";
const SHOTS = "/tmp/rc_shots";

let pass = 0, fail = 0;
const ok = (n) => { pass++; console.log(`  \x1b[32mPASS\x1b[0m ${n}`); };
const bad = (n, e) => { fail++; console.log(`  \x1b[31mFAIL\x1b[0m ${n}${e ? " — " + e : ""}`); };

const b = await chromium.launch();
const page = await b.newPage();
page.on("dialog", (d) => d.accept());
const victim = `delui_${Date.now()}@rentcar.com`;
try {
  await page.goto(ADMIN_APP, { waitUntil: "networkidle" });
  await page.getByPlaceholder("ime@primer.com").fill("admin@rentcar.com");
  await page.getByPlaceholder("••••••••").fill("admin12345");
  await page.getByRole("button", { name: "Prijavi se" }).click();
  await page.getByRole("heading", { name: "Administracija" }).waitFor({ timeout: 10000 });
  ok("admin ulogovan u admin app");

  // create a victim account
  await page.getByPlaceholder("ime@primer.com").fill(victim);
  await page.getByPlaceholder("Najmanje 6 karaktera").fill("password123");
  await page.getByRole("button", { name: "Kreiraj nalog" }).click();
  await page.getByText(new RegExp(`Nalog ${victim} je kreiran`)).waitFor({ timeout: 8000 });
  ok(`napravljen nalog za brisanje: ${victim}`);

  // own row must NOT have a delete button (self-protection in UI)
  const ownRow = page.locator("li").filter({ hasText: "admin@rentcar.com" }).first();
  await ownRow.waitFor({ timeout: 5000 });
  (await ownRow.getByRole("button", { name: "Obriši" }).count()) === 0
    ? ok("admin NEMA 'Obriši' na svom redu (self-protection)")
    : bad("admin ima delete na sebi!");

  // delete the victim
  const row = page.locator("li").filter({ hasText: victim });
  await row.first().getByRole("button", { name: "Obriši" }).click();
  await page.getByText(new RegExp(`Nalog ${victim} je obrisan`)).waitFor({ timeout: 8000 });
  ok("nalog obrisan (poruka prikazana)");

  // it must disappear from the list
  await page.locator("li").filter({ hasText: victim }).first().waitFor({ state: "detached", timeout: 5000 });
  ok("obrisani nalog nestao iz liste");
  await page.screenshot({ path: `${SHOTS}/delete-after.png`, fullPage: true });
} catch (e) {
  bad("neočekivana greška", e.message);
  await page.screenshot({ path: `${SHOTS}/delete-error.png` }).catch(() => {});
} finally {
  await b.close();
}
console.log(`\n REZULTAT: \x1b[32m${pass} prošlo\x1b[0m, \x1b[31m${fail} palo\x1b[0m`);
process.exit(fail);
