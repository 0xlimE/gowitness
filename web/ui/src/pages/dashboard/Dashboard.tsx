import { useEffect, useState } from "react";
import { WideSkeleton } from "@/components/loading";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { GlobeIcon, LayersIcon, ServerIcon, BuildingIcon } from "lucide-react";
import { Bar, BarChart, CartesianGrid, XAxis, YAxis, ResponsiveContainer } from "recharts";
import { ChartContainer, ChartLegend, ChartLegendContent, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import { DomainExplorer } from "@/components/domain-explorer";
import * as apitypes from "@/lib/api/types";
import { getData } from "./data";

const chartConfig = {
  count: {
    label: "Total",
    color: "hsl(var(--chart-5))",
  },
  code: {
    label: "HTTP Status Code",
    color: "hsl(var(--chart-1))",
  },
} satisfies ChartConfig;

const domainChartConfig = {
  count: {
    label: "Count",
    color: "hsl(var(--chart-3))",
  },
  domain: {
    label: "Domain",
    color: "hsl(var(--chart-2))",
  },
} satisfies ChartConfig;

const StatCard = ({ title, value, icon: Icon }: { title: string; value: number | string; icon: React.ElementType; }) => (
  <Card className="overflow-hidden transition-all hover:shadow-lg">
    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
      <CardTitle className="text-sm font-medium">{title}</CardTitle>
      <Icon className="h-4 w-4 text-muted-foreground" />
    </CardHeader>
    <CardContent>
      <div className="text-2xl font-bold">{value}</div>
    </CardContent>
  </Card>
);

export default function DashboardPage() {
  const [stats, setStats] = useState<apitypes.statistics>();
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    getData(setLoading, setStats);
  }, []);

  if (loading) return <WideSkeleton />;

  return (
    <div className="space-y-8">
      <h1 className="text-3xl font-bold tracking-tight">
        {stats?.target_info ? `${stats.target_info.company_name} Assessment` : "Dashboard"}
      </h1>
      
      {/* Target Information Card */}
      {stats?.target_info && (
        <Card className="bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BuildingIcon className="h-5 w-5" />
              Target Information
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col md:flex-row gap-6 items-center md:items-start">
              {/* Info Section */}
              <div className="flex-1 grid gap-4 md:grid-cols-3">
                <div>
                  <div className="text-sm font-medium text-muted-foreground">Company</div>
                  <div className="text-lg font-semibold">{stats.target_info.company_name}</div>
                </div>
                <div>
                  <div className="text-sm font-medium text-muted-foreground">Main Domain</div>
                  <div className="text-lg font-semibold">{stats.target_info.main_domain}</div>
                </div>
                <div>
                  <div className="text-sm font-medium text-muted-foreground">Scan Started</div>
                  <div className="text-lg font-semibold">{new Date(stats.target_info.scan_start_time).toLocaleDateString()}</div>
                </div>
              </div>
              
              {/* Logo Section - Now on the right */}
              <div className="flex-shrink-0">
                <div className="w-32 h-32 bg-white rounded-lg shadow-md p-3 flex items-center justify-center">
                  <img 
                    src="/api/logo" 
                    alt={`${stats.target_info.company_name} logo`}
                    className="max-w-full max-h-full object-contain"
                    onError={(e) => {
                      // Replace with placeholder icon if logo fails to load
                      const img = e.target as HTMLImageElement;
                      img.style.display = 'none';
                      const parent = img.parentElement;
                      if (parent) {
                        parent.innerHTML = '<div class="text-muted-foreground"><svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg></div>';
                      }
                    }}
                  />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <StatCard
          title="Unique Apex Domains"
          value={stats?.domain_stats ? stats.domain_stats.unique_apex_domains : 0}
          icon={GlobeIcon}
        />
        <StatCard
          title="Total Subdomains"
          value={stats?.domain_stats ? stats.domain_stats.total_subdomains : 0}
          icon={LayersIcon}
        />
        <StatCard
          title="Unique IP Addresses"
          value={stats?.ip_stats ? stats.ip_stats.unique_ips : 0}
          icon={ServerIcon}
        />
      </div>
      
      <div className="grid gap-4 grid-cols-1 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Top Apex Domains by Count</CardTitle>
            </CardHeader>
            <CardContent>
              <ChartContainer config={domainChartConfig} className="aspect-auto h-[350px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={stats?.domain_stats?.apex_domains?.slice(0, 10).map(domain => ({
                    domain: domain.domain,
                    count: domain.count
                  })) || []}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="domain"
                      tickLine={false}
                      axisLine={false}
                      angle={-45}
                      textAnchor="end"
                      height={80}
                    />
                    <YAxis
                      tickLine={false}
                      axisLine={false}
                    />
                    <ChartTooltip content={<ChartTooltipContent hideLabel indicator="line" />} />
                    <ChartLegend content={<ChartLegendContent />} />
                    <Bar dataKey="count" fill="var(--color-count)" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </ChartContainer>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>HTTP Status Code Distribution</CardTitle>
            </CardHeader>
            <CardContent>
              <ChartContainer config={chartConfig} className="aspect-auto h-[350px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={stats?.response_code_stats}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="code"
                      tickLine={false}
                      axisLine={false}
                    />
                    <YAxis
                      tickLine={false}
                      axisLine={false}
                    />
                    <ChartTooltip content={<ChartTooltipContent hideLabel indicator="line" />} />
                    <ChartLegend content={<ChartLegendContent />} />
                    <Bar dataKey="count" fill="var(--color-count)" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </ChartContainer>
            </CardContent>
          </Card>
        </div>
        
        <div className="lg:col-span-1 space-y-4">
          <DomainExplorer domains={stats?.domain_stats?.apex_domains || []} />
        </div>
      </div>
    </div>
  );
}