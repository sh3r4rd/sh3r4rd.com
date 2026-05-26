import { useEffect, useMemo, useState } from "react";
import { Search } from "lucide-react";
import { Button } from "../ui/button";

const inputClass =
  "p-2 border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100";

export default function FilterBar({ data, filters, onFilterChange }) {
  const [searchInput, setSearchInput] = useState(filters.search);

  // Debounce search input changes (200ms) before pushing up to parent.
  useEffect(() => {
    const t = setTimeout(() => {
      if (searchInput !== filters.search) {
        onFilterChange({ ...filters, search: searchInput });
      }
    }, 200);
    return () => clearTimeout(t);
    // Debounce should react only to local input changes, not to filter resets
    // pushed in from the parent (e.g. Clear Filters). The sync effect below
    // handles those cases.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchInput]);

  // If parent resets search externally (e.g. Clear Filters), sync local state.
  useEffect(() => {
    if (filters.search !== searchInput) {
      setSearchInput(filters.search);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters.search]);

  const companyOptions = useMemo(
    () => [...new Set((data || []).map((r) => r.company).filter(Boolean))].sort(),
    [data],
  );
  const jobTitleOptions = useMemo(
    () => [...new Set((data || []).map((r) => r.jobTitle).filter(Boolean))].sort(),
    [data],
  );

  const hasActiveFilters = Boolean(
    filters.search ||
      filters.company ||
      filters.jobTitle ||
      filters.monthFrom ||
      filters.monthTo,
  );

  const handleClear = () => {
    onFilterChange({
      search: "",
      company: "",
      jobTitle: "",
      monthFrom: "",
      monthTo: "",
    });
  };

  return (
    <section className="space-y-3">
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <div className="relative">
          <Search
            className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500 dark:text-gray-400 pointer-events-none"
            aria-hidden="true"
          />
          <input
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search company or job title…"
            aria-label="Search company or job title"
            className={`${inputClass} pl-9 w-full`}
          />
        </div>

        <select
          value={filters.company}
          onChange={(e) =>
            onFilterChange({ ...filters, company: e.target.value })
          }
          aria-label="Filter by company"
          className={`${inputClass} w-full`}
        >
          <option value="">All companies</option>
          {companyOptions.map((c) => (
            <option key={c} value={c}>
              {c}
            </option>
          ))}
        </select>

        <select
          value={filters.jobTitle}
          onChange={(e) =>
            onFilterChange({ ...filters, jobTitle: e.target.value })
          }
          aria-label="Filter by job title"
          className={`${inputClass} w-full`}
        >
          <option value="">All job titles</option>
          {jobTitleOptions.map((j) => (
            <option key={j} value={j}>
              {j}
            </option>
          ))}
        </select>

        {hasActiveFilters && (
          <div className="flex items-center">
            <Button size="sm" onClick={handleClear} type="button">
              Clear Filters
            </Button>
          </div>
        )}
      </div>

      <div className="flex flex-wrap gap-3 items-end">
        <label className="flex flex-col text-sm text-gray-700 dark:text-gray-300">
          <span className="mb-1">From</span>
          <input
            type="month"
            value={filters.monthFrom}
            onChange={(e) =>
              onFilterChange({ ...filters, monthFrom: e.target.value })
            }
            aria-label="From month"
            className={inputClass}
          />
        </label>
        <label className="flex flex-col text-sm text-gray-700 dark:text-gray-300">
          <span className="mb-1">To</span>
          <input
            type="month"
            value={filters.monthTo}
            onChange={(e) =>
              onFilterChange({ ...filters, monthTo: e.target.value })
            }
            aria-label="To month"
            className={inputClass}
          />
        </label>
      </div>
    </section>
  );
}
