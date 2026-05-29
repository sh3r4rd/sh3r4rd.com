import { useMemo, useState } from "react";
import { ArrowUpDown, ArrowUp, ArrowDown, Inbox } from "lucide-react";
import { Card, CardContent } from "../ui/card";

const COLUMNS = [
  { key: "company", label: "Company" },
  { key: "jobTitle", label: "Job Title" },
  { key: "recruiterLabel", label: "Recruiter" },
  { key: "month", label: "Month" },
];

const pageBtnBase = "px-3 py-1 text-sm rounded-md";
const pageBtnIdle = `${pageBtnBase} border border-gray-300 dark:border-gray-700 bg-white text-gray-700 dark:bg-gray-800 dark:text-gray-300`;

function SortIcon({ active, direction }) {
  if (!active) {
    return <ArrowUpDown className="w-4 h-4" aria-hidden="true" />;
  }
  return direction === "asc" ? (
    <ArrowUp className="w-4 h-4" aria-hidden="true" />
  ) : (
    <ArrowDown className="w-4 h-4" aria-hidden="true" />
  );
}

function getPageNumbers(page, totalPages) {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, i) => i + 1);
  }
  const candidates = [1, page - 1, page, page + 1, totalPages].filter(
    (n) => n >= 1 && n <= totalPages
  );
  const unique = Array.from(new Set(candidates)).sort((a, b) => a - b);
  const result = [];
  for (let i = 0; i < unique.length; i++) {
    if (i > 0 && unique[i] - unique[i - 1] > 1) {
      result.push("…");
    }
    result.push(unique[i]);
  }
  return result;
}

export default function RecruiterTable({
  data = [],
  page,
  pageSize = 10,
  onPageChange,
}) {
  const [sort, setSort] = useState({ key: "month", direction: "desc" });

  const sortedData = useMemo(() => {
    const copy = [...data];
    copy.sort((a, b) => {
      const av = a[sort.key];
      const bv = b[sort.key];
      const aEmpty = av === undefined || av === null || av === "";
      const bEmpty = bv === undefined || bv === null || bv === "";
      if (aEmpty && bEmpty) return 0;
      if (aEmpty) return 1;
      if (bEmpty) return -1;
      const cmp =
        typeof av === "number" && typeof bv === "number"
          ? av - bv
          : String(av).localeCompare(String(bv), undefined, {
              sensitivity: "base",
            });
      return sort.direction === "asc" ? cmp : -cmp;
    });
    return copy;
  }, [data, sort]);

  const totalPages = Math.max(1, Math.ceil(sortedData.length / pageSize));
  const currentPage = Math.max(1, Math.min(page, totalPages));
  const startIdx = (currentPage - 1) * pageSize;
  const endIdx = Math.min(currentPage * pageSize, sortedData.length);
  const pageRows = sortedData.slice(startIdx, endIdx);

  function handleSort(key) {
    setSort((prev) => {
      if (prev.key !== key) {
        return { key, direction: "asc" };
      }
      return { key, direction: prev.direction === "asc" ? "desc" : "asc" };
    });
  }

  if (data.length === 0) {
    return (
      <div className="text-center py-12">
        <Inbox className="w-12 h-12 mx-auto text-gray-400 dark:text-gray-600" />
        <p className="mt-4 text-gray-600 dark:text-gray-400">
          No recruiter requests match your filters
        </p>
      </div>
    );
  }

  const pageNumbers = getPageNumbers(currentPage, totalPages);

  return (
    <div>
      {/* Desktop table */}
      <div className="hidden md:block overflow-x-auto">
        <table className="min-w-full">
          <thead className="bg-gray-50 dark:bg-gray-800 text-gray-700 dark:text-gray-300 text-left text-sm font-medium uppercase tracking-wider">
            <tr>
              {COLUMNS.map((col) => {
                const active = sort.key === col.key;
                return (
                  <th
                    key={col.key}
                    scope="col"
                    aria-sort={
                      active
                        ? sort.direction === "asc"
                          ? "ascending"
                          : "descending"
                        : "none"
                    }
                    className="p-0"
                  >
                    <button
                      type="button"
                      onClick={() => handleSort(col.key)}
                      className="flex w-full items-center gap-2 px-4 py-3 text-left cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 select-none"
                    >
                      {col.label}
                      <SortIcon active={active} direction={sort.direction} />
                    </button>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {pageRows.map((r) => (
              <tr
                key={r.id}
                className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800"
              >
                <td className="px-4 py-3 text-sm text-gray-900 dark:text-gray-100">
                  {r.company || "—"}
                </td>
                <td className="px-4 py-3 text-sm text-gray-900 dark:text-gray-100">
                  {r.jobTitle || "—"}
                </td>
                <td className="px-4 py-3 text-sm text-gray-900 dark:text-gray-100">
                  {r.recruiterLabel || "Anonymous"}
                </td>
                <td className="px-4 py-3 text-sm text-gray-900 dark:text-gray-100">
                  {r.month || "—"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Mobile cards */}
      <div className="md:hidden space-y-3">
        {pageRows.map((r) => (
          <Card key={r.id}>
            <CardContent>
              <div className="text-lg font-bold text-gray-900 dark:text-gray-100">
                {r.company || "—"}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">
                {r.jobTitle || "—"}
              </div>
              <div className="mt-2 text-sm text-gray-700 dark:text-gray-300">
                {r.recruiterLabel || "Anonymous"}
              </div>
              <div className="mt-1 text-xs text-gray-500 dark:text-gray-400 text-right">
                {r.month || "—"}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between flex-wrap gap-3">
          <div className="text-sm text-gray-600 dark:text-gray-400">
            Showing {startIdx + 1}-{endIdx} of {sortedData.length}
          </div>
          <div className="flex items-center gap-1 flex-wrap">
            <button
              type="button"
              onClick={() => onPageChange(currentPage - 1)}
              disabled={currentPage === 1}
              className={`${pageBtnIdle} ${
                currentPage === 1 ? "opacity-50 cursor-not-allowed" : ""
              }`}
            >
              Previous
            </button>
            {pageNumbers.map((n, i) =>
              n === "…" ? (
                <span
                  key={`ellipsis-${i}`}
                  className="px-2 text-sm text-gray-500 dark:text-gray-400"
                >
                  …
                </span>
              ) : (
                <button
                  key={n}
                  type="button"
                  onClick={() => onPageChange(n)}
                  className={
                    n === currentPage
                      ? `${pageBtnBase} bg-blue-600 text-white`
                      : pageBtnIdle
                  }
                >
                  {n}
                </button>
              )
            )}
            <button
              type="button"
              onClick={() => onPageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
              className={`${pageBtnIdle} ${
                currentPage === totalPages ? "opacity-50 cursor-not-allowed" : ""
              }`}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
