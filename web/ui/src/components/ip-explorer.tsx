import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { ChevronDownIcon, ChevronRightIcon, ServerIcon, GlobeIcon, InfoIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import * as apitypes from "@/lib/api/types";
import { IPDetailView } from "./ip-detail-view";

interface IPExplorerProps {
  ipStats: apitypes.ip_statistics;
}

interface GroupedDomain {
  domain: string;
  entries: Array<{
    result_id: number;
    url: string;
    protocol: string;
    port: string;
  }>;
}

export function IPExplorer({ ipStats }: IPExplorerProps) {
  const [openIPs, setOpenIPs] = useState<Set<string>>(new Set());
  const [openDomains, setOpenDomains] = useState<Set<string>>(new Set());
  const [selectedIP, setSelectedIP] = useState<string | null>(null);

  // If an IP is selected, show the detailed view
  if (selectedIP) {
    return <IPDetailView ipAddress={selectedIP} onBack={() => setSelectedIP(null)} />;
  }

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
    // Navigate to the screenshot page for this result
    window.location.href = `/screenshot/${resultId}`;
  };

  // Group domains by hostname for each IP
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

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <ServerIcon className="h-5 w-5" />
          IP Address Explorer
          <Badge variant="secondary">{ipStats.unique_ips} unique IPs</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {ipStats.ip_list.map((ip) => (
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
                      setSelectedIP(ip.ip_address);
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
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
