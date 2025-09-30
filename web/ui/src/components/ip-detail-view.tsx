import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { 
  ChevronDownIcon, 
  ChevronRightIcon, 
  ServerIcon, 
  GlobeIcon, 
  MapPinIcon,
  NetworkIcon,
  InfoIcon,
  BuildingIcon,
  WifiIcon,
  HashIcon,
  MonitorIcon,
  ClockIcon,
  TagIcon,
  AlertTriangleIcon,
  MapIcon,
  CompassIcon,
  MailIcon,
  RouteIcon
} from "lucide-react";
import { get } from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";

interface IPDetailViewProps {
  ipAddress: string;
  onBack: () => void;
}

export function IPDetailView({ ipAddress, onBack }: IPDetailViewProps) {
  const [ipInfo, setIpInfo] = useState<apitypes.IPInfoResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [openSections, setOpenSections] = useState<Set<string>>(new Set(['ports', 'domains']));

  useEffect(() => {
    const fetchIPInfo = async () => {
      try {
        setLoading(true);
        const response = await get('ipinfo', { ip: ipAddress });
        setIpInfo(response);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load IP information');
      } finally {
        setLoading(false);
      }
    };

    fetchIPInfo();
  }, [ipAddress]);

  const toggleSection = (section: string) => {
    const newOpenSections = new Set(openSections);
    if (newOpenSections.has(section)) {
      newOpenSections.delete(section);
    } else {
      newOpenSections.add(section);
    }
    setOpenSections(newOpenSections);
  };

  if (loading) {
    return (
      <Card className="w-full">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <ServerIcon className="h-5 w-5" />
              IP Address Details
            </CardTitle>
            <Button onClick={onBack} variant="outline" size="sm">
              ← Back to IP List
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
            <span className="ml-2">Loading IP information...</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="w-full">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <ServerIcon className="h-5 w-5" />
              IP Address Details
            </CardTitle>
            <Button onClick={onBack} variant="outline" size="sm">
              ← Back to IP List
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="text-red-600 py-4">
            Error: {error}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!ipInfo) {
    return (
      <Card className="w-full">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <ServerIcon className="h-5 w-5" />
              IP Address Details
            </CardTitle>
            <Button onClick={onBack} variant="outline" size="sm">
              ← Back to IP List
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="py-4">No information available for this IP address.</div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <ServerIcon className="h-5 w-5" />
            {ipInfo.ip_address}
          </CardTitle>
          <Button onClick={onBack} variant="outline" size="sm">
            ← Back to IP List
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        
        {/* Quick Summary */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
            <div className="flex items-center gap-2">
              <NetworkIcon className="h-5 w-5 text-blue-600" />
              <div>
                <div className="text-2xl font-bold text-blue-600">{ipInfo.total_ports}</div>
                <div className="text-sm text-muted-foreground">Open Ports</div>
              </div>
            </div>
          </div>
          
          <div className="bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 rounded-lg p-4">
            <div className="flex items-center gap-2">
              <GlobeIcon className="h-5 w-5 text-green-600" />
              <div>
                <div className="text-2xl font-bold text-green-600">{ipInfo.total_domains}</div>
                <div className="text-sm text-muted-foreground">Domains</div>
              </div>
            </div>
          </div>
          
          <div className="bg-purple-50 dark:bg-purple-950/20 border border-purple-200 dark:border-purple-800 rounded-lg p-4">
            <div className="flex items-center gap-2">
              <InfoIcon className="h-5 w-5 text-purple-600" />
              <div>
                <div className="text-2xl font-bold text-purple-600">
                  {ipInfo.shodan_info?.vulns?.length || 0}
                </div>
                <div className="text-sm text-muted-foreground">Vulnerabilities</div>
              </div>
            </div>
          </div>
          
          <div className="bg-orange-50 dark:bg-orange-950/20 border border-orange-200 dark:border-orange-800 rounded-lg p-4">
            <div className="flex items-center gap-2">
              <TagIcon className="h-5 w-5 text-orange-600" />
              <div>
                <div className="text-2xl font-bold text-orange-600">
                  {ipInfo.shodan_info?.tags?.length || 0}
                </div>
                <div className="text-sm text-muted-foreground">Tags</div>
              </div>
            </div>
          </div>
        </div>
        
        {/* Shodan Information */}
        {ipInfo.shodan_info && (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <InfoIcon className="h-5 w-5" />
              Shodan Intelligence
            </h3>
            
            {/* Organization & ISP Information */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {ipInfo.shodan_info.organization && (
                <div className="bg-card border rounded-lg p-4">
                  <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                    <BuildingIcon className="h-4 w-4" />
                    Organization
                  </div>
                  <div className="font-semibold text-foreground">{ipInfo.shodan_info.organization}</div>
                </div>
              )}
              
              {ipInfo.shodan_info.isp && (
                <div className="bg-card border rounded-lg p-4">
                  <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                    <WifiIcon className="h-4 w-4" />
                    ISP
                  </div>
                  <div className="font-semibold text-foreground">{ipInfo.shodan_info.isp}</div>
                </div>
              )}
              
              {ipInfo.shodan_info.asn && (
                <div className="bg-card border rounded-lg p-4">
                  <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                    <HashIcon className="h-4 w-4" />
                    ASN
                  </div>
                  <div className="font-semibold text-foreground">{ipInfo.shodan_info.asn}</div>
                </div>
              )}
            </div>

            {/* Location Information */}
            {(ipInfo.shodan_info.country || ipInfo.shodan_info.city || ipInfo.shodan_info.region) && (
              <div className="space-y-3">
                <h4 className="text-md font-semibold flex items-center gap-2">
                  <MapPinIcon className="h-4 w-4" />
                  Geographic Information
                </h4>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  {ipInfo.shodan_info.country && (
                    <div className="bg-card border rounded-lg p-4">
                      <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                        <MapIcon className="h-4 w-4" />
                        Country
                      </div>
                      <div className="font-semibold text-foreground">
                        {ipInfo.shodan_info.country}
                        {ipInfo.shodan_info.country_code && (
                          <span className="text-muted-foreground ml-1">({ipInfo.shodan_info.country_code})</span>
                        )}
                      </div>
                    </div>
                  )}

                  {ipInfo.shodan_info.city && (
                    <div className="bg-card border rounded-lg p-4">
                      <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                        <BuildingIcon className="h-4 w-4" />
                        City
                      </div>
                      <div className="font-semibold text-foreground">{ipInfo.shodan_info.city}</div>
                    </div>
                  )}

                  {ipInfo.shodan_info.region && (
                    <div className="bg-card border rounded-lg p-4">
                      <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                        <CompassIcon className="h-4 w-4" />
                        Region
                      </div>
                      <div className="font-semibold text-foreground">{ipInfo.shodan_info.region}</div>
                    </div>
                  )}

                  {ipInfo.shodan_info.postal && (
                    <div className="bg-card border rounded-lg p-4">
                      <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                        <MailIcon className="h-4 w-4" />
                        Postal Code
                      </div>
                      <div className="font-semibold text-foreground">{ipInfo.shodan_info.postal}</div>
                    </div>
                  )}
                </div>

                {/* Coordinates */}
                {(ipInfo.shodan_info.latitude !== undefined && ipInfo.shodan_info.longitude !== undefined) && (
                  <div className="bg-card border rounded-lg p-4">
                    <div className="font-medium text-sm text-muted-foreground mb-2 flex items-center gap-2">
                      <RouteIcon className="h-4 w-4" />
                      Coordinates
                    </div>
                    <div className="font-mono text-sm">
                      <span className="font-semibold">Lat:</span> {ipInfo.shodan_info.latitude}, 
                      <span className="font-semibold ml-2">Lng:</span> {ipInfo.shodan_info.longitude}
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Technical Information */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {ipInfo.shodan_info.os && (
                <div className="bg-card border rounded-lg p-4">
                  <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                    <MonitorIcon className="h-4 w-4" />
                    Operating System
                  </div>
                  <div className="font-semibold text-foreground">{ipInfo.shodan_info.os}</div>
                </div>
              )}
              
              {(ipInfo.shodan_info.last_update || ipInfo.shodan_info.updated_at) && (
                <div className="bg-card border rounded-lg p-4">
                  <div className="font-medium text-sm text-muted-foreground mb-1 flex items-center gap-2">
                    <ClockIcon className="h-4 w-4" />
                    Last Updated
                  </div>
                  <div className="font-semibold text-foreground">
                    {ipInfo.shodan_info.last_update || ipInfo.shodan_info.updated_at}
                  </div>
                </div>
              )}
            </div>

            {/* Ports from Shodan */}
            {ipInfo.shodan_info.ports && ipInfo.shodan_info.ports.length > 0 && (
              <div className="space-y-3">
                <h4 className="text-md font-semibold flex items-center gap-2">
                  <NetworkIcon className="h-4 w-4" />
                  Shodan Discovered Ports ({ipInfo.shodan_info.ports.length})
                </h4>
                <div className="flex flex-wrap gap-2">
                  {ipInfo.shodan_info.ports.map((port, index) => (
                    <Badge key={index} variant="outline" className="font-mono">
                      {port}
                    </Badge>
                  ))}
                </div>
              </div>
            )}

            {/* Tags */}
            {ipInfo.shodan_info.tags && ipInfo.shodan_info.tags.length > 0 && (
              <div className="space-y-3">
                <h4 className="text-md font-semibold flex items-center gap-2">
                  <TagIcon className="h-4 w-4" />
                  Tags
                </h4>
                <div className="flex flex-wrap gap-2">
                  {ipInfo.shodan_info.tags.map((tag, index) => (
                    <Badge key={index} variant="secondary">{tag}</Badge>
                  ))}
                </div>
              </div>
            )}

            {/* Vulnerabilities */}
            {ipInfo.shodan_info.vulns && ipInfo.shodan_info.vulns.length > 0 && (
              <div className="space-y-3">
                <h4 className="text-md font-semibold flex items-center gap-2 text-destructive">
                  <AlertTriangleIcon className="h-4 w-4" />
                  Vulnerabilities ({ipInfo.shodan_info.vulns.length})
                </h4>
                <div className="flex flex-wrap gap-2">
                  {ipInfo.shodan_info.vulns.map((vuln, index) => (
                    <Badge key={index} variant="destructive">{vuln}</Badge>
                  ))}
                </div>
              </div>
            )}

            {/* Hostnames */}
            {ipInfo.shodan_info.hostnames && ipInfo.shodan_info.hostnames.length > 0 && (
              <div className="space-y-3">
                <h4 className="text-md font-semibold flex items-center gap-2">
                  <GlobeIcon className="h-4 w-4" />
                  Reverse DNS Hostnames
                </h4>
                <div className="flex flex-wrap gap-2">
                  {ipInfo.shodan_info.hostnames.map((hostname, index) => (
                    <Badge key={index} variant="outline" className="font-mono">{hostname}</Badge>
                  ))}
                </div>
              </div>
            )}

            {/* Shodan Domains */}
            {ipInfo.shodan_info.shodan_domains && ipInfo.shodan_info.shodan_domains.length > 0 && (
              <div className="space-y-3">
                <h4 className="text-md font-semibold flex items-center gap-2">
                  <GlobeIcon className="h-4 w-4" />
                  Shodan Discovered Domains
                </h4>
                <div className="flex flex-wrap gap-2">
                  {ipInfo.shodan_info.shodan_domains.map((domain, index) => (
                    <Badge key={index} variant="outline" className="font-mono text-blue-600">
                      {domain}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Open Ports */}
        <Collapsible 
          open={openSections.has('ports')} 
          onOpenChange={() => toggleSection('ports')}
        >
          <CollapsibleTrigger className="flex items-center gap-2 text-lg font-semibold hover:text-blue-600 transition-colors">
            {openSections.has('ports') ? (
              <ChevronDownIcon className="h-5 w-5" />
            ) : (
              <ChevronRightIcon className="h-5 w-5" />
            )}
            <NetworkIcon className="h-5 w-5" />
            Open Ports ({ipInfo.total_ports})
          </CollapsibleTrigger>
          <CollapsibleContent className="mt-4">
            {ipInfo.open_ports.length === 0 ? (
              <div className="text-muted-foreground italic">No open ports found</div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {ipInfo.open_ports.map((port) => (
                  <div key={port.id} className="bg-card border rounded-lg p-4 hover:shadow-md transition-shadow">
                    <div className="flex items-center justify-between mb-3">
                      <div className="font-semibold text-lg flex items-center gap-2">
                        <NetworkIcon className="h-4 w-4 text-blue-500" />
                        {port.port}/{port.protocol.toLowerCase()}
                      </div>
                      <Badge 
                        variant={port.state === 'open' ? 'default' : 'secondary'}
                        className={port.state === 'open' ? 'bg-green-500' : ''}
                      >
                        {port.state}
                      </Badge>
                    </div>
                    
                    {port.service && (
                      <div className="mb-2">
                        <div className="text-xs text-muted-foreground mb-1">Service</div>
                        <div className="text-sm font-medium">{port.service}</div>
                      </div>
                    )}
                    
                    <div className="flex flex-wrap gap-2 mb-3">
                      {port.is_cdn && (
                        <Badge variant="outline" className="text-xs">
                          CDN: {port.cdn_name || 'Unknown'}
                        </Badge>
                      )}
                      {port.protocol && (
                        <Badge variant="secondary" className="text-xs">
                          {port.protocol.toUpperCase()}
                        </Badge>
                      )}
                    </div>
                    
                    {port.banner && (
                      <details className="mt-2">
                        <summary className="text-xs text-muted-foreground cursor-pointer hover:text-foreground">
                          Show Banner
                        </summary>
                        <div className="text-xs font-mono bg-muted p-2 rounded mt-2 max-h-24 overflow-y-auto">
                          {port.banner}
                        </div>
                      </details>
                    )}
                    
                    <div className="text-xs text-muted-foreground mt-2">
                      Discovered: {new Date(port.discovered_at).toLocaleDateString()}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CollapsibleContent>
        </Collapsible>

        {/* Associated Domains */}
        <Collapsible 
          open={openSections.has('domains')} 
          onOpenChange={() => toggleSection('domains')}
        >
          <CollapsibleTrigger className="flex items-center gap-2 text-lg font-semibold hover:text-blue-600 transition-colors">
            {openSections.has('domains') ? (
              <ChevronDownIcon className="h-5 w-5" />
            ) : (
              <ChevronRightIcon className="h-5 w-5" />
            )}
            <GlobeIcon className="h-5 w-5" />
            Associated Domains ({ipInfo.total_domains})
          </CollapsibleTrigger>
          <CollapsibleContent className="mt-4">
            {ipInfo.domains.length === 0 ? (
              <div className="text-muted-foreground italic">No associated domains found</div>
            ) : (
              <div className="space-y-3">
                {ipInfo.domains.map((domain) => (
                  <div 
                    key={domain.id} 
                    className="bg-card border rounded-lg p-4 hover:shadow-md cursor-pointer transition-all hover:border-blue-300"
                    onClick={() => window.location.href = `/screenshot/${domain.id}`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="font-semibold text-blue-600 hover:underline flex items-center gap-2">
                          <GlobeIcon className="h-4 w-4" />
                          {domain.url}
                        </div>
                        {domain.title && (
                          <div className="text-sm text-muted-foreground mt-1">{domain.title}</div>
                        )}
                      </div>
                      <div className="text-right flex flex-col items-end gap-2">
                        <Badge 
                          variant={domain.response_code < 400 ? 'default' : 'destructive'}
                          className={domain.response_code < 400 ? 'bg-green-500' : ''}
                        >
                          HTTP {domain.response_code}
                        </Badge>
                        <Badge variant="secondary" className="text-xs">
                          {domain.protocol.toUpperCase()}
                        </Badge>
                      </div>
                    </div>
                    
                    {domain.failed && (
                      <div className="mt-2 p-2 bg-destructive/10 border border-destructive/20 rounded">
                        <div className="text-xs text-destructive flex items-center gap-1">
                          <AlertTriangleIcon className="h-3 w-3" />
                          Failed: {domain.failed_reason}
                        </div>
                      </div>
                    )}
                    
                    <div className="flex items-center justify-between mt-3 text-xs text-muted-foreground">
                      <div className="flex items-center gap-1">
                        <ClockIcon className="h-3 w-3" />
                        Probed: {new Date(domain.probed_at).toLocaleString()}
                      </div>
                      <div className="flex items-center gap-1">
                        <InfoIcon className="h-3 w-3" />
                        ID: {domain.id}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CollapsibleContent>
        </Collapsible>
      </CardContent>
    </Card>
  );
}
