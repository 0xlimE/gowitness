// stats
type statistics = {
  dbsize: number;
  results: number;
  headers: number;
  consolelogs: number;
  networklogs: number;
  response_code_stats: response_code_stats[];
  domain_stats: domain_statistics;
  ip_stats: ip_statistics;
  target_info?: target_information;
};

interface target_information {
  company_name: string;
  main_domain: string;
  scan_start_time: string;
  scan_status: string;
  notes: string;
}

interface response_code_stats {
  code: number;
  count: number;
}

interface domain_statistics {
  unique_apex_domains: number;
  total_subdomains: number;
  total_domains: number;
  apex_domains: apex_domain[];
}

interface ip_statistics {
  unique_ips: number;
  total_results: number;
  ip_list: ip_entry[];
}

interface apex_domain {
  domain: string;
  is_apex: boolean;
  result_id?: number;
  subdomains: subdomain[];
  count: number;
}

interface subdomain {
  domain: string;
  result_id: number;
  url: string;
  protocol: string;
  port: string;
}

interface ip_entry {
  ip_address: string;
  domain_count: number;
  first_seen: string;
  last_seen: string;
  sample_domain: string;
  result_id: number;
  domains: ip_domain_entry[];
}

interface ip_domain_entry {
  domain: string;
  result_id: number;
  url: string;
  protocol: string;
  port: string;
}

// wappalyzer
type wappalyzer = {
  [name: string]: string;
};

// gallery
type gallery = {
  results: galleryResult[];
  page: number;
  limit: number;
  total_count: number;
};

type galleryResult = {
  id: number;
  url: string;
  probed_at: string;
  title: string;
  response_code: number;
  file_name: string;
  screenshot: string;
  failed: boolean;
  technologies: string[];
};

// list
type list = {
  id: number;
  url: string;
  final_url: string;
  response_code: number;
  response_reason: string;
  protocol: string;
  content_length: number;
  title: string;
  failed: boolean;
  failed_reason: string;
};

// details
interface tls {
  id: number;
  result_id: number;
  protocol: string;
  key_exchange: string;
  cipher: string;
  subject_name: string;
  san_list: sanlist[];
  issuer: string;
  valid_from: string;
  valid_to: string;
  server_signature_algorithm: number;
  encrypted_client_hello: boolean;
}

interface sanlist {
  id: number;
  tls_id: number;
  value: string;
}

interface technology {
  id: number;
  result_id: number;
  value: string;
}

interface header {
  id: number;
  result_id: number;
  key: string;
  value: string | null;
}

interface networklog {
  id: number;
  result_id: number;
  request_type: number;
  status_code: number;
  url: string;
  remote_ip: string;
  mime_type: string;
  time: string;
  error: string;
  content: string;
}

interface consolelog {
  id: number;
  resultid: number;
  type: string;
  value: string;
}

interface cookie {
  id: number;
  result_id: number;
  name: string;
  value: string;
  domain: string;
  path: string;
  expires: string; // actually a timestamp
  size: number;
  http_only: boolean;
  secure: boolean;
  session: boolean;
  priority: string;
  source_scheme: string;
  source_port: number;
}

interface detail {
  id: number;
  url: string;
  ip_address: string;
  probed_at: string;
  final_url: string;
  response_code: number;
  response_reason: string;
  protocol: string;
  content_length: number;
  html: string;
  title: string;
  perception_hash: string;
  file_name: string;
  is_pdf: boolean;
  failed: boolean;
  failed_reason: string;
  screenshot: string;
  tls: tls;
  technologies: technology[];
  headers: header[];
  network: networklog[];
  console: consolelog[];
  cookies: cookie[];
}

interface searchresult {
  id: number;
  url: string;
  final_url: string;
  response_code: number;
  content_length: number;
  title: string;
  matched_fields: string[];
  file_name: string;
  screenshot: string;
}

interface technologylist {
  technologies: string[];
}

// IP information with Shodan data
interface IPPortInfo {
  id: number;
  port: number;
  protocol: string;
  service: string;
  state: string;
  banner: string;
  scan_session_id?: number;
  discovered_at: string;
  is_cdn: boolean;
  cdn_name: string;
  cdn_detected: boolean;
  original_host: string;
}

interface DomainInfo {
  id: number;
  url: string;
  final_url: string;
  title: string;
  response_code: number;
  response_reason: string;
  protocol: string;
  screenshot: string;
  file_name: string;
  failed: boolean;
  failed_reason: string;
  probed_at: string;
  scan_session_id?: number;
}

interface ShodanInfo {
  organization?: string;
  isp?: string;
  asn?: string;
  country?: string;
  country_code?: string;
  city?: string;
  region?: string;
  postal?: string;
  latitude?: number;
  longitude?: number;
  os?: string;
  tags?: string[];
  ports?: number[];
  hostnames?: string[];
  shodan_domains?: string[];
  vulns?: string[];
  last_update?: string;
  updated_at?: string;
}

interface IPInfoResponse {
  ip_address: string;
  open_ports: IPPortInfo[];
  total_ports: number;
  domains: DomainInfo[];
  total_domains: number;
  scan_sessions: number[];
  shodan_info?: ShodanInfo;
}

export type {
  statistics,
  wappalyzer,
  gallery,
  list,
  galleryResult,
  tls,
  sanlist,
  technology,
  header,
  networklog,
  consolelog,
  cookie,
  detail,
  searchresult,
  technologylist,
  domain_statistics,
  apex_domain,
  subdomain,
  ip_statistics,
  ip_entry,
  ip_domain_entry,
  target_information,
  IPPortInfo,
  DomainInfo,
  ShodanInfo,
  IPInfoResponse,
};