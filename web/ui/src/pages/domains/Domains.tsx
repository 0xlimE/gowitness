import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { ChevronDownIcon, ChevronRightIcon, GlobeIcon, LayersIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { WideSkeleton } from "@/components/loading";
import * as apitypes from "@/lib/api/types";
import * as api from "@/lib/api/api";
import { toast } from "@/hooks/use-toast";

interface GroupedSubdomain {
  domain: string;
  entries: Array<{
    result_id: number;
    url: string;
    protocol: string;
    port: string;
  }>;
}

export default function DomainsPage() {
  const [stats, setStats] = useState<apitypes.statistics>();
  const [loading, setLoading] = useState<boolean>(true);
  const [openDomains, setOpenDomains] = useState<Set<string>>(new Set());
  const [openSubdomains, setOpenSubdomains] = useState<Set<string>>(new Set());
  const navigate = useNavigate();

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const s = await api.get('statistics');
        setStats(s);
      } catch (err) {
        toast({
          title: "API Error",
          variant: "destructive",
          description: `Failed to get statistics: ${err}`
        });
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const toggleDomain = (domain: string) => {
    const newOpenDomains = new Set(openDomains);
    if (newOpenDomains.has(domain)) {
      newOpenDomains.delete(domain);
    } else {
      newOpenDomains.add(domain);
    }
    setOpenDomains(newOpenDomains);
  };

  const toggleSubdomain = (subdomainKey: string) => {
    const newOpenSubdomains = new Set(openSubdomains);
    if (newOpenSubdomains.has(subdomainKey)) {
      newOpenSubdomains.delete(subdomainKey);
    } else {
      newOpenSubdomains.add(subdomainKey);
    }
    setOpenSubdomains(newOpenSubdomains);
  };

  const handleEntryClick = (resultId: number) => {
    navigate(`/screenshot/${resultId}`);
  };

  const handleApexClick = (resultId?: number) => {
    if (resultId) {
      navigate(`/screenshot/${resultId}`);
    }
  };

  const groupSubdomains = (subdomains: apitypes.subdomain[]): GroupedSubdomain[] => {
    const grouped = new Map<string, GroupedSubdomain>();
    
    subdomains.forEach(sub => {
      if (!grouped.has(sub.domain)) {
        grouped.set(sub.domain, {
          domain: sub.domain,
          entries: []
        });
      }
      grouped.get(sub.domain)!.entries.push({
        result_id: sub.result_id,
        url: sub.url,
        protocol: sub.protocol,
        port: sub.port
      });
    });

    return Array.from(grouped.values()).sort((a, b) => a.domain.localeCompare(b.domain));
  };

  if (loading) return <WideSkeleton />;

  const domainStats = stats?.domain_stats;
  const sortedDomains = domainStats?.apex_domains 
    ? [...domainStats.apex_domains].sort((a, b) => b.count - a.count)
    : [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Domain Browser</h1>
        <p className="text-muted-foreground mt-2">
          Explore all discovered domains organized by apex domain. Expand to see subdomains and their variants.
        </p>
      </div>

      {domainStats && (
        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Apex Domains</CardTitle>
              <GlobeIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{domainStats.unique_apex_domains}</div>
              <p className="text-xs text-muted-foreground">Unique apex domains discovered</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Subdomains</CardTitle>
              <LayersIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{domainStats.total_subdomains}</div>
              <p className="text-xs text-muted-foreground">Total subdomains across all domains</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Domains</CardTitle>
              <GlobeIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{domainStats.total_domains}</div>
              <p className="text-xs text-muted-foreground">All domains including variants</p>
            </CardContent>
          </Card>
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <GlobeIcon className="h-5 w-5" />
            Apex Domains
          </CardTitle>
          <CardDescription>
            Domains are organized by apex domain and sorted by screenshot count. 
            Click on an apex domain to view its screenshot, or expand to see all subdomains.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {sortedDomains.length > 0 ? (
              sortedDomains.map((domain) => (
                <Collapsible key={domain.domain} open={openDomains.has(domain.domain)}>
                  <CollapsibleTrigger
                    className="flex items-center justify-between w-full px-3 py-2 text-left bg-muted rounded-lg hover:bg-muted/80 transition-colors"
                    onClick={() => toggleDomain(domain.domain)}
                  >
                    <div className="flex items-center gap-2">
                      {openDomains.has(domain.domain) ? (
                        <ChevronDownIcon className="h-4 w-4" />
                      ) : (
                        <ChevronRightIcon className="h-4 w-4" />
                      )}
                      <span
                        className="font-medium cursor-pointer hover:text-blue-600"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleApexClick(domain.result_id);
                        }}
                      >
                        {domain.domain}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{domain.count} screenshots</Badge>
                      <LayersIcon className="h-4 w-4 text-muted-foreground" />
                    </div>
                  </CollapsibleTrigger>
                  
                  <CollapsibleContent className="mt-2 ml-6 space-y-1">
                    {groupSubdomains(domain.subdomains || []).map((groupedSub) => (
                      <Collapsible key={groupedSub.domain} open={openSubdomains.has(`${domain.domain}-${groupedSub.domain}`)}>
                        <CollapsibleTrigger
                          className="flex items-center justify-between w-full px-3 py-2 text-left bg-background border rounded hover:bg-muted/50 transition-colors"
                          onClick={() => toggleSubdomain(`${domain.domain}-${groupedSub.domain}`)}
                        >
                          <div className="flex items-center gap-2">
                            {openSubdomains.has(`${domain.domain}-${groupedSub.domain}`) ? (
                              <ChevronDownIcon className="h-3 w-3" />
                            ) : (
                              <ChevronRightIcon className="h-3 w-3" />
                            )}
                            <span className="text-sm text-muted-foreground">{groupedSub.domain}</span>
                          </div>
                          <Badge variant="secondary" className="text-xs">
                            {groupedSub.entries.length} variants
                          </Badge>
                        </CollapsibleTrigger>
                        
                        <CollapsibleContent className="mt-1 ml-4 space-y-1">
                          {groupedSub.entries.map((entry) => (
                            <div
                              key={entry.result_id}
                              className="px-3 py-2 text-sm bg-card border rounded cursor-pointer hover:bg-accent transition-colors"
                              onClick={() => handleEntryClick(entry.result_id)}
                            >
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                  <span className="text-xs font-mono text-blue-600">{entry.protocol}</span>
                                  <span className="text-xs text-muted-foreground">:</span>
                                  <span className="text-xs font-mono text-green-600">{entry.port}</span>
                                </div>
                                <Badge variant="outline" className="text-xs">
                                  {entry.result_id}
                                </Badge>
                              </div>
                            </div>
                          ))}
                        </CollapsibleContent>
                      </Collapsible>
                    ))}
                  </CollapsibleContent>
                </Collapsible>
              ))
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                No domains found in the database.
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
