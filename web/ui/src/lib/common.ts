import { toast } from "@/hooks/use-toast";
import * as apitypes from "@/lib/api/types";

const copyToClipboard = (content: string, type: string) => {
  navigator.clipboard.writeText(content).then(() => {
    toast({
      description: `${type} copied to clipboard`,
    });
  }).catch((err) => {
    console.error('Failed to copy content: ', err);
    toast({
      title: "Error",
      description: "Failed to copy content",
      variant: "destructive",
    });
  });
};

const getIconUrl = (tech: string, wappalyzer: apitypes.wappalyzer | undefined): string | undefined => {
  if (!wappalyzer || !(tech in wappalyzer)) return undefined;

  return wappalyzer[tech];
};

const getStatusColor = (code: number) => {
  if (code >= 200 && code < 300) return "bg-green-500 text-white";
  if (code >= 400 && code < 500) return "bg-yellow-500 text-black";
  if (code >= 500) return "bg-red-500 text-white";
  return "bg-gray-500 text-white";
};

// Extract IP address from a detail object
// First check if ip_address is available from the API response
// If not, try to extract from URL if it's a direct IP
const extractIPAddress = (detail: apitypes.detail): string | null => {
  // First, check if the API response includes the ip_address field
  if (detail.ip_address && detail.ip_address.trim() !== '') {
    return detail.ip_address.trim();
  }

  // If not available from API, try to extract from URL if it's a direct IP
  try {
    const url = new URL(detail.final_url || detail.url);
    const hostname = url.hostname;
    
    // Check if hostname is an IPv4 address using regex
    const ipv4Regex = /^(\d{1,3}\.){3}\d{1,3}$/;
    // Check if hostname is an IPv6 address (simplified check)
    const ipv6Regex = /^([0-9a-f]{0,4}:){2,7}[0-9a-f]{0,4}$/i;
    
    if (ipv4Regex.test(hostname) || ipv6Regex.test(hostname)) {
      return hostname;
    }
  } catch (error) {
    // Invalid URL, ignore
  }

  return null;
};

export { copyToClipboard, getIconUrl, getStatusColor, extractIPAddress };