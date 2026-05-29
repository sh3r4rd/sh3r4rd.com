import { Mail, Building2, Briefcase, CalendarDays } from "lucide-react";
import { Card, CardContent } from "../ui/card";

function pickTopJobTitle(topJobTitles) {
  let top = null;
  let topCount = 0;
  for (const [title, count] of Object.entries(topJobTitles)) {
    if (count > topCount) {
      top = title;
      topCount = count;
    }
  }
  return top ?? "N/A";
}

export default function StatsCards({ stats }) {
  const s = stats ?? {};
  const byMonth = s.byMonth ?? {};
  const topJobTitles = s.topJobTitles ?? {};

  // Current month as a UTC "YYYY-MM" key. Note: the server's byMonth keys are
  // derived from date_day[:7], which the email-parser formats from the email's
  // Date header in its *original* timezone (not UTC). So for emails whose Date
  // header straddles a month boundary relative to UTC, this lookup can under- or
  // over-count "This Month". The robust fix is to format date_day in UTC server
  // side; until then this is a known low-severity edge case.
  const currentMonth = new Date().toISOString().slice(0, 7);

  const totalRequests = s.totalEmails ?? 0;
  const uniqueCompanies = s.uniqueCompanies ?? 0;
  const topJobTitle = pickTopJobTitle(topJobTitles);
  const thisMonthCount = byMonth[currentMonth] ?? 0;

  const cards = [
    {
      label: "Total Requests",
      value: totalRequests,
      icon: <Mail className="w-5 h-5 text-blue-600 dark:text-blue-300" />,
      iconBg: "bg-blue-100 dark:bg-blue-900/40",
    },
    {
      label: "Unique Companies",
      value: uniqueCompanies,
      icon: <Building2 className="w-5 h-5 text-green-600 dark:text-green-300" />,
      iconBg: "bg-green-100 dark:bg-green-900/40",
    },
    {
      label: "Top Job Title",
      value: topJobTitle,
      icon: <Briefcase className="w-5 h-5 text-purple-600 dark:text-purple-300" />,
      iconBg: "bg-purple-100 dark:bg-purple-900/40",
    },
    {
      label: "This Month",
      value: thisMonthCount,
      icon: <CalendarDays className="w-5 h-5 text-orange-600 dark:text-orange-300" />,
      iconBg: "bg-orange-100 dark:bg-orange-900/40",
    },
  ];

  return (
    <section>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {cards.map(({ label, value, icon, iconBg }) => (
          <Card key={label}>
            <CardContent>
              <div className="flex items-center gap-3">
                <div className={`rounded-full p-2 ${iconBg}`}>
                  {icon}
                </div>
                <div className="min-w-0">
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {label}
                  </p>
                  <p
                    className="text-2xl font-bold text-gray-900 dark:text-white truncate"
                    title={typeof value === "string" ? value : undefined}
                  >
                    {value}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
