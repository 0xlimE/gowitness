import { gallery, list, statistics, wappalyzer, detail, searchresult, technologylist, IPInfoResponse } from "@/lib/api/types";

// Dynamically determine the base API path from the current URL
function getApiBasePath(): string {
  const pathname = window.location.pathname;
  // If we're at /project/something/, use that as base for API calls
  const projectMatch = pathname.match(/^(\/project\/[^\/]+)\//);
  if (projectMatch) {
    return projectMatch[1] + `/api`;
  }
  // Otherwise use root
  return `/api`;
}

// Dynamically determine the screenshots base path
function getScreenshotsBasePath(): string {
  const pathname = window.location.pathname;
  // If we're at /project/something/, use that as base for screenshots
  const projectMatch = pathname.match(/^(\/project\/[^\/]+)\//);
  if (projectMatch) {
    return projectMatch[1] + `/screenshots`;
  }
  // Otherwise use root
  return `/screenshots`;
}

const endpoints = {
  // screenshot path (kept for backward compatibility)
  screenshot: {
    path: `/screenshots`,
    returnas: [] // n/a
  },

  // get endpoints
  statistics: {
    path: `/statistics`,
    returnas: {} as statistics
  },
  wappalyzer: {
    path: `/wappalyzer`,
    returnas: {} as wappalyzer
  },
  gallery: {
    path: `/results/gallery`,
    returnas: {} as gallery
  },
  list: {
    path: `/results/list`,
    returnas: [] as list[]
  },
  detail: {
    path: `/results/detail/:id`,
    returnas: {} as detail
  },
  technology: {
    path: `/results/technology`,
    returnas: {} as technologylist
  },
  ipinfo: {
    path: `/ip/:ip`,
    returnas: {} as IPInfoResponse
  },

  // post endpoints
  search: {
    path: `/search`,
    returnas: {} as searchresult
  },
  delete: {
    path: `/results/delete`,
    returnas: "" as string
  },
  submit: {
    path: `/submit`,
    returnas: "" as string
  },
  submitsingle: {
    path: `/submit/single`,
    returnas: {} as detail
  }
};

type Endpoints = typeof endpoints;
type EndpointReturnType<K extends keyof Endpoints> = Endpoints[K]['returnas'];

const replacePathParams = (path: string, params?: Record<string, string | number | boolean>): [string, Record<string, string | number | boolean>] => {
  if (!params) return [path, {}];

  const paramRegex = /:([a-zA-Z0-9_]+)/g;
  const missingParams: string[] = [];
  const remainingParams = { ...params }; // Create a copy of the params object to modify

  // Replace all `:param` placeholders with the corresponding values from params
  const newPath = path.replace(paramRegex, (match, paramName) => {
    if (paramName in remainingParams) {
      const value = remainingParams[paramName];
      delete remainingParams[paramName];
      return encodeURIComponent(value.toString());
    } else {
      missingParams.push(paramName);
      return match;
    }
  });

  // If any required params were missing, throw an error
  if (missingParams.length > 0) {
    throw new Error(`Missing required parameters: ${missingParams.join(', ')}`);
  }

  return [newPath, remainingParams];
};

const serializeParams = (params: Record<string, string | number | boolean>) => {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    query.append(key, value.toString());
  });
  return query.toString() ? `?${query.toString()}` : '';
};

const get = async <K extends keyof Endpoints>(
  endpointKey: K,
  params?: Record<string, string | number | boolean>,
  raw: boolean = false
): Promise<EndpointReturnType<K>> => {

  const endpoint = endpoints[endpointKey];
  const [pathWithParams, remainingParams] = replacePathParams(endpoint.path, params);
  const queryString = remainingParams ? serializeParams(remainingParams) : '';

  // Dynamically determine the base API path for each request
  const basePath = import.meta.env.VITE_GOWITNESS_API_BASE_URL 
    ? import.meta.env.VITE_GOWITNESS_API_BASE_URL + `/api`
    : getApiBasePath();

  const res = await fetch(`${basePath}${pathWithParams}${queryString}`);

  if (!res.ok) throw new Error(`HTTP Error: ${res.status}`);

  if (raw) return await res.text() as unknown as EndpointReturnType<K>;
  return await res.json() as EndpointReturnType<K>;
};

const post = async <K extends keyof Endpoints>(
  endpointKey: K,
  data?: unknown
): Promise<EndpointReturnType<K>> => {

  const endpoint = endpoints[endpointKey];
  
  // Dynamically determine the base API path for each request
  const basePath = import.meta.env.VITE_GOWITNESS_API_BASE_URL 
    ? import.meta.env.VITE_GOWITNESS_API_BASE_URL + `/api`
    : getApiBasePath();

  const res = await fetch(`${basePath}${endpoint.path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });

  if (!res.ok) throw new Error(`HTTP Error: ${res.status}`);

  return await res.json() as EndpointReturnType<K>;
};

// Export the screenshot path function for use in components
const getScreenshotUrl = (filename: string): string => {
  const basePath = import.meta.env.VITE_GOWITNESS_API_BASE_URL 
    ? import.meta.env.VITE_GOWITNESS_API_BASE_URL + `/screenshots`
    : getScreenshotsBasePath();
  return `${basePath}/${filename}`;
};

export { endpoints, get, post, getScreenshotUrl };