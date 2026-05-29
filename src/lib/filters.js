// Shared empty/default state for the recruiter dashboard filters.
// Imported by both DashboardPage (initial state / reset) and FilterBar
// (default prop) so the filter shape has a single source of truth.
export const EMPTY_FILTERS = {
  search: "",
  company: "",
  jobTitle: "",
  monthFrom: "",
  monthTo: "",
};
