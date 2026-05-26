import { useCallback, useEffect, useMemo, useState } from "react";
import { AlertCircle, Loader2, RefreshCw } from "lucide-react";
import { Button } from "../components/ui/button";
import Breadcrumbs from "../components/layout/Breadcrumbs";
import Header from "../components/layout/Header";
import StatsCards from "../components/sections/StatsCards";
import FilterBar from "../components/sections/FilterBar";
import RecruiterTable from "../components/sections/RecruiterTable";

const API_URL = "https://api.sh3r4rd.com/recruiters";
const PAGE_SIZE = 10;

const EMPTY_FILTERS = {
  search: "",
  company: "",
  jobTitle: "",
  monthFrom: "",
  monthTo: "",
};

function matchesFilters(item, filters) {
  const { search, company, jobTitle, monthFrom, monthTo } = filters;

  if (company && item.company !== company) return false;
  if (jobTitle && item.jobTitle !== jobTitle) return false;

  if (search) {
    const needle = search.toLowerCase();
    const haystack = `${item.company ?? ""} ${item.jobTitle ?? ""}`.toLowerCase();
    if (!haystack.includes(needle)) return false;
  }

  if (monthFrom && (item.month ?? "") < monthFrom) return false;
  if (monthTo && (item.month ?? "") > monthTo) return false;

  return true;
}

export default function DashboardPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filters, setFilters] = useState(EMPTY_FILTERS);
  const [page, setPage] = useState(1);
  const [refreshing, setRefreshing] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(API_URL);
      if (!res.ok) {
        throw new Error(`Request failed: ${res.status}`);
      }
      const body = await res.json();
      setData(Array.isArray(body) ? body : []);
    } catch (err) {
      setError(err.message || "Failed to load recruiter data");
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const filteredData = useMemo(
    () => data.filter((item) => matchesFilters(item, filters)),
    [data, filters],
  );

  const handleFilterChange = useCallback((nextFilters) => {
    setFilters(nextFilters);
    setPage(1);
  }, []);

  const handleRefresh = () => {
    setRefreshing(true);
    loadData();
  };

  const filtersActive = useMemo(
    () => Object.values(filters).some((v) => v !== ""),
    [filters],
  );

  return (
    <section className="max-w-4xl mx-auto space-y-12">
      <Breadcrumbs />
      <Header />

      <section className="space-y-6">
        <div className="flex items-center justify-between flex-wrap gap-3">
          <h2 className="text-2xl font-semibold">Recruiter Dashboard</h2>
          <button
            type="button"
            onClick={handleRefresh}
            disabled={loading || refreshing}
            className="inline-flex items-center gap-2 px-3 py-2 text-sm rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <RefreshCw
              className={`w-4 h-4 ${refreshing ? "animate-spin" : ""}`}
            />
            Refresh
          </button>
        </div>

        {loading && !refreshing ? (
          <div className="flex flex-col items-center justify-center py-16 text-gray-600 dark:text-gray-400">
            <Loader2 className="w-8 h-8 animate-spin" />
            <p className="mt-4">Loading recruiter data...</p>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <AlertCircle className="w-10 h-10 text-red-500 dark:text-red-400" />
            <p className="mt-4 text-gray-700 dark:text-gray-300">{error}</p>
            <div className="mt-4">
              <Button onClick={handleRefresh}>Try Again</Button>
            </div>
          </div>
        ) : (
          <>
            <StatsCards data={data} />

            <FilterBar
              data={data}
              filters={filters}
              onFilterChange={handleFilterChange}
            />

            <p className="text-sm text-gray-600 dark:text-gray-400">
              {filtersActive
                ? `${filteredData.length} results (filtered from ${data.length})`
                : `${data.length} results`}
            </p>

            <RecruiterTable
              data={filteredData}
              page={page}
              pageSize={PAGE_SIZE}
              onPageChange={setPage}
            />
          </>
        )}
      </section>
    </section>
  );
}
