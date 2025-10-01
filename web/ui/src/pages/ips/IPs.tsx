import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { ChevronDownIcon, ChevronRightIcon, ServerIcon, GlobeIcon, InfoIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { WideSkeleton } from "@/components/loading";
import * as apitypes from "@/lib/api/types";
import * as api from "@/lib/api/api";
import { toast } from "@/hooks/use-toast";

interface GroupedDomain {
  domain: string;
  entries: Array<{
    result_id: number;
    url: string;
    protocol: string;
    port: string;
  }>;
}

export default function IPsPage() {
  const [stats, setStats] = useState<apitypes.statistics>();
  const [loading, setLoading] = useState<boolean>(true);
  const [openIPs, setOpenIPs] = useState<Set<string>>(new Set());
  const [openDomains, setOpenDomains] = useState<Set<string>>(new Set());
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

  const toggleIP = (ip: string) => {
    const newOpenIPs = new Set(openIPs);
    if (newOpenIPs.has(ip)) {
      newOpenIPs.delete(ip);
    } else {
      newOpenIPs.add(ip);
    }
    setOpenIPs(newOpenIPs);
  };

  const toggleDomain = (domainKey: string) => {
    const newOpenDomains = new Set(openDomains);
    if (newOpenDomains.has(domainKey)) {
      newOpenDomains.delete(domainKey);
    } else {
      newOpenDomains.add(domainKey);
    }
    setOpenDomains(newOpenDomains);
  };

  const handleEntryClick = (resultId: number) => {
    navigate(`/screenshot/${resultId}`);
  };

  const handleIPClick = (ipAddress: string) => {
    navigate(`/ip/${ipAddress}`);
  };

  const groupDomains = (domains: apitypes.ip_domain_entry[]): GroupedDomain[] => {
    const grouped = new Map<string, GroupedDomain>();
    
    domains.forEach(domain => {
      if (!grouped.has(domain.domain)) {
        grouped.set(domain.domain, {
          domain: domain.domain,
          entries: []
        });
      }
      grouped.get(domain.domain)!.entries.push({
        result_id: domain.result_id,
        url: domain.url,
        protocol: domain.protocol,
        port: domain.port
      });
    });

    return Array.from(grouped.values()).sort((a, b) => a.domain.localeCompare(b.domain));
  };

  if (loading) return <WideSkeleton />;

  const ipStats = stats?.ip_stats;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">IP Address Browser</h1>
        <p className="text-muted-foreground mt-2">
          Explore all discovered IP addresses and their associated domains. Click on an IP address for detailed information.
        </p>
      </div>

      {ipStats && (
        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total IPs</CardTitle>
              <ServerIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{ipStats.unique_ips}</div>
              <p className="text-xs text-muted-foreground">Unique IP addresses discovered</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Results</CardTitle>
              <GlobeIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{ipStats.total_results}</div>
              <p className="text-xs text-muted-foreground">Screenshots across all IPs</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Avg Domains/IP</CardTitle>
              <InfoIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {ipStats.unique_ips > 0 ? Math.round(ipStats.total_results / ipStats.unique_ips) : 0}
              </div>
              <p className="text-xs text-muted-foreground">Average domains per IP</p>
            </CardContent>
          </Card>
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ServerIcon className="h-5 w-5" />
            IP Addresses
          </CardTitle>
          <CardDescription>
            Browse by IP address. Click the "Details" button to view comprehensive information about an IP,
            or expand to see all domains and their variants.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {ipStats?.ip_list && ipStats.ip_list.length > 0 ? (
              ipStats.ip_list.map((ip) => (
                <Collapsible key={ip.ip_address} open={openIPs.has(ip.ip_address)}>
                  <CollapsibleTrigger
                    className="flex items-center justify-between w-full px-3 py-2 text-left bg-muted rounded-lg hover:bg-muted/80 transition-colors"
                    onClick={() => toggleIP(ip.ip_address)}
                  >
                    <div className="flex items-center gap-2">
                      {openIPs.has(ip.ip_address) ? (
                        <ChevronDownIcon className="h-4 w-4" />
                      ) : (
                        <ChevronRightIcon className="h-4 w-4" />
                      )}
                      <span className="font-mono text-sm">{ip.ip_address}</span>
                      <span className="text-xs text-muted-foreground">({ip.sample_domain})</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{ip.domain_count} domains</Badge>
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        onClick={(e) => {
                          e.stopPropagation();
                          handleIPClick(ip.ip_address);
                        }}
                        className="px-2 py-1 h-auto"
                      >
                        <InfoIcon className="h-3 w-3 mr-1" />
                        Details
                      </Button>
                      <ServerIcon className="h-4 w-4 text-muted-foreground" />
                    </div>
                  </CollapsibleTrigger>
                  
                  <CollapsibleContent className="mt-2 ml-6 space-y-1">
                    {groupDomains(ip.domains || []).map((groupedDomain) => (
                      <Collapsible key={groupedDomain.domain} open={openDomains.has(`${ip.ip_address}-${groupedDomain.domain}`)}>
                        <CollapsibleTrigger
                          className="flex items-center justify-between w-full px-3 py-2 text-left bg-background border rounded hover:bg-muted/50 transition-colors"
                          onClick={() => toggleDomain(`${ip.ip_address}-${groupedDomain.domain}`)}
                        >
                          <div className="flex items-center gap-2">
                            {openDomains.has(`${ip.ip_address}-${groupedDomain.domain}`) ? (
                              <ChevronDownIcon className="h-3 w-3" />
                            ) : (
                              <ChevronRightIcon className="h-3 w-3" />
                            )}
                            <GlobeIcon className="h-3 w-3 text-muted-foreground" />
                            <span className="text-sm">{groupedDomain.domain}</span>
                          </div>
                          <Badge variant="secondary" className="text-xs">
                            {groupedDomain.entries.length} variants
                          </Badge>
                        </CollapsibleTrigger>
                        
                        <CollapsibleContent className="mt-1 ml-4 space-y-1">
                          {groupedDomain.entries.map((entry) => (
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
                No IP addresses found in the database.
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
