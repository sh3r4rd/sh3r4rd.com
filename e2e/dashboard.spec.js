import { test, expect } from "@playwright/test";

const RECRUITERS_URL = "https://api.sh3r4rd.com/recruiters";
const STATS_URL = "https://api.sh3r4rd.com/stats";

/**
 * Deterministic fixture: 12 recruiters spanning 3 companies, several job titles,
 * and 3 months. >10 rows exercises pagination (PAGE_SIZE = 10). Months are fixed
 * (never "current") so the data is stable regardless of when the suite runs.
 */
const RECRUITERS = [
  { id: "1", company: "Acme Corp", jobTitle: "Frontend Engineer", month: "2025-01", recruiterLabel: "Alice", confidence: 0.9 },
  { id: "2", company: "Acme Corp", jobTitle: "Backend Engineer", month: "2025-01", recruiterLabel: "Bob", confidence: 0.8 },
  { id: "3", company: "Acme Corp", jobTitle: "DevOps Engineer", month: "2025-02", recruiterLabel: "Carol", confidence: 0.7 },
  { id: "4", company: "Globex", jobTitle: "Frontend Engineer", month: "2025-02", recruiterLabel: "Dave", confidence: 0.95 },
  { id: "5", company: "Globex", jobTitle: "Data Scientist", month: "2025-02", recruiterLabel: "Eve", confidence: 0.6 },
  { id: "6", company: "Globex", jobTitle: "Product Manager", month: "2025-03", recruiterLabel: "Frank", confidence: 0.5 },
  { id: "7", company: "Initech", jobTitle: "Backend Engineer", month: "2025-03", recruiterLabel: "Grace", confidence: 0.85 },
  { id: "8", company: "Initech", jobTitle: "Frontend Engineer", month: "2025-03", recruiterLabel: "Heidi", confidence: 0.75 },
  { id: "9", company: "Initech", jobTitle: "DevOps Engineer", month: "2025-01", recruiterLabel: "Ivan", confidence: 0.65 },
  { id: "10", company: "Initech", jobTitle: "Data Scientist", month: "2025-02", recruiterLabel: "Judy", confidence: 0.55 },
  { id: "11", company: "Acme Corp", jobTitle: "Product Manager", month: "2025-03", recruiterLabel: "Mallory", confidence: 0.45 },
  { id: "12", company: "Globex", jobTitle: "Backend Engineer", month: "2025-01", recruiterLabel: "Niaj", confidence: 0.35 },
];

const STATS = {
  totalEmails: 12,
  uniqueCompanies: 3,
  byMonth: { "2025-01": 4, "2025-02": 4, "2025-03": 4 },
  topJobTitles: { "Frontend Engineer": 3, "Backend Engineer": 3, "DevOps Engineer": 2 },
};

/**
 * Register API mocks for BOTH endpoints the dashboard fetches on load. Must be
 * called BEFORE page.goto so neither request escapes to the network.
 */
async function mockApi(page, { recruiters = RECRUITERS, recruitersStatus = 200, stats = STATS } = {}) {
  await page.route(RECRUITERS_URL, (route) =>
    route.fulfill({
      status: recruitersStatus,
      contentType: "application/json",
      body: JSON.stringify(recruiters),
    }),
  );
  await page.route(STATS_URL, (route) =>
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(stats),
    }),
  );
}

/** Locate a stats card by its label and return the locator for its value <p>. */
function statValue(page, label) {
  return page
    .locator("p", { hasText: new RegExp(`^${label}$`) })
    .locator("xpath=following-sibling::p[1]");
}

test.describe("Recruiter Dashboard", () => {
  test("loads and renders heading with table rows", async ({ page }) => {
    await mockApi(page);
    await page.goto("/dashboard");

    await expect(
      page.getByRole("heading", { name: "Recruiter Dashboard" }),
    ).toBeVisible();

    // Results count line reflects the full dataset (no filter).
    await expect(page.getByText("12 results", { exact: true })).toBeVisible();
    await expect(page.getByRole("button", { name: "Refresh" })).toBeVisible();

    // First page shows PAGE_SIZE (10) rows; pagination summary confirms total.
    await expect(page.getByText("Showing 1-10 of 12")).toBeVisible();
  });

  test("stats cards show totals", async ({ page }) => {
    await mockApi(page);
    await page.goto("/dashboard");

    await expect(statValue(page, "Total Requests")).toHaveText("12");
    await expect(statValue(page, "Unique Companies")).toHaveText("3");
    // Top job title is the max-count entry in topJobTitles.
    await expect(statValue(page, "Top Job Title")).toHaveText("Frontend Engineer");
  });

  test("search filter narrows results", async ({ page }) => {
    await mockApi(page);
    await page.goto("/dashboard");
    await expect(page.getByText("12 results", { exact: true })).toBeVisible();

    // "data scientist" matches company+jobTitle substring; 2 rows (Globex, Initech).
    await page
      .getByLabel("Search company or job title")
      .fill("data scientist");

    await expect(
      page.getByText("2 results (filtered from 12)"),
    ).toBeVisible();
    // Row-level cell assertion is desktop-only: the <table> is display:none
    // below the md breakpoint, so its cells leave the accessibility tree.
    if (page.viewportSize().width >= 768) {
      await expect(
        page.getByRole("cell", { name: "Data Scientist" }),
      ).toHaveCount(2);
    }
  });

  test("company filter narrows results", async ({ page }) => {
    await mockApi(page);
    await page.goto("/dashboard");
    await expect(page.getByText("12 results", { exact: true })).toBeVisible();

    // Initech has 4 recruiters.
    await page.getByLabel("Filter by company").selectOption("Initech");

    await expect(
      page.getByText("4 results (filtered from 12)"),
    ).toBeVisible();
    // Desktop-only: see note in the search-filter test.
    if (page.viewportSize().width >= 768) {
      await expect(page.getByRole("cell", { name: "Initech" })).toHaveCount(4);
    }
  });

  test("clear filters restores full dataset", async ({ page }) => {
    await mockApi(page);
    await page.goto("/dashboard");

    await page.getByLabel("Filter by company").selectOption("Globex");
    await expect(
      page.getByText("4 results (filtered from 12)"),
    ).toBeVisible();

    const clear = page.getByRole("button", { name: "Clear Filters" });
    await expect(clear).toBeVisible();
    await clear.click();

    await expect(page.getByText("12 results", { exact: true })).toBeVisible();
    await expect(clear).toBeHidden();
  });

  test("sort by column and toggle direction", async ({ page }) => {
    // Sortable column headers exist only in the desktop table; the mobile card
    // layout has no sortable headers, so this journey is desktop-only.
    test.skip(
      page.viewportSize().width < 768,
      "sortable table headers are desktop-only",
    );
    await mockApi(page);
    await page.goto("/dashboard");
    await expect(page.getByText("12 results", { exact: true })).toBeVisible();

    const companyHeader = page.getByRole("columnheader", { name: "Company" });
    const companyHeaderBtn = companyHeader.getByRole("button", {
      name: "Company",
    });

    // Default sort is month descending; Company starts unsorted.
    await expect(companyHeader).toHaveAttribute("aria-sort", "none");

    // First click -> ascending.
    await companyHeaderBtn.click();
    await expect(companyHeader).toHaveAttribute("aria-sort", "ascending");
    // Ascending: first company alphabetically is "Acme Corp".
    await expect(
      page.locator("tbody tr").first().locator("td").first(),
    ).toHaveText("Acme Corp");

    // Second click toggles -> descending.
    await companyHeaderBtn.click();
    await expect(companyHeader).toHaveAttribute("aria-sort", "descending");
    // Descending: first company is "Initech".
    await expect(
      page.locator("tbody tr").first().locator("td").first(),
    ).toHaveText("Initech");
  });

  test("empty state when API returns no recruiters", async ({ page }) => {
    await mockApi(page, { recruiters: [] });
    await page.goto("/dashboard");

    await expect(
      page.getByText("No recruiter requests match your filters"),
    ).toBeVisible();
    await expect(page.getByText("0 results", { exact: true })).toBeVisible();
  });

  test("error state when recruiters API returns 500", async ({ page }) => {
    await mockApi(page, { recruitersStatus: 500 });
    await page.goto("/dashboard");

    await expect(page.getByText("Request failed: 500")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Try Again" }),
    ).toBeVisible();
    // Table should not render in the error state.
    await expect(page.locator("table")).toHaveCount(0);
  });
});

test.describe("Mobile layout", () => {
  // Only meaningful at the mobile viewport (< 768px md breakpoint).
  test.skip(
    ({ viewport }) => !viewport || viewport.width >= 768,
    "mobile-only assertions",
  );

  test("renders card layout instead of table at 375px", async ({ page }) => {
    await mockApi(page);
    await page.goto("/dashboard");
    await expect(page.getByText("12 results", { exact: true })).toBeVisible();

    // The desktop <table> is hidden below the md breakpoint.
    await expect(page.locator("table")).toBeHidden();

    // Card content (a company name) is visible instead. Filter to visible
    // matches so we don't latch onto the hidden <option> in the company
    // dropdown, which also contains the text "Acme Corp".
    await expect(
      page
        .getByText("Acme Corp", { exact: true })
        .filter({ visible: true })
        .first(),
    ).toBeVisible();
  });
});

test.describe("Dark mode", () => {
  test.use({ colorScheme: "dark" });

  test("uses dark background", async ({ page }) => {
    await mockApi(page);
    await page.goto("/dashboard");
    await expect(
      page.getByRole("heading", { name: "Recruiter Dashboard" }),
    ).toBeVisible();

    // <main> root: bg-white / dark:bg-gray-900 (rgb(17, 24, 39)).
    await expect(page.locator("main")).toHaveCSS(
      "background-color",
      "rgb(17, 24, 39)",
    );
  });
});
