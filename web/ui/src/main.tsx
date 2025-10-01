import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { createBrowserRouter, RouterProvider, redirect } from "react-router-dom";
import '@/index.css';

import App from '@/pages/App';
import ErrorPage from '@/pages/Error';

import DashboardPage from '@/pages/dashboard/Dashboard';
import GalleryPage from '@/pages/gallery/Gallery';
import TablePage from '@/pages/table/Table';
import ScreenshotDetailPage from '@/pages/detail/Detail';
import SearchResultsPage from '@/pages/search/Search';
import IPPage from '@/pages/ip/IP';
import IPsPage from '@/pages/ips/IPs';
import DomainsPage from '@/pages/domains/Domains';
import IntroPage from '@/pages/intro/Intro';

import { searchAction } from '@/pages/search/action';
import { searchLoader } from '@/pages/search/loader';
import { deleteAction } from '@/pages/detail/actions';
import { hasSeenIntro } from '@/lib/cookies';

// Dynamically determine the base path from the current URL
function getBasename(): string {
  const pathname = window.location.pathname;
  // If we're at /project/something/, use that as basename
  const projectMatch = pathname.match(/^(\/project\/[^\/]+)\//);
  if (projectMatch) {
    return projectMatch[1];
  }
  // Otherwise use root
  return '/';
}

const router = createBrowserRouter([
  {
    path: '/intro',
    element: <IntroPage />,
  },
  {
    path: '/',
    element: <App />,
    errorElement: <ErrorPage />,
    children: [
      {
        path: '/',
        element: <DashboardPage />,
        loader: () => {
          // Redirect to intro if user hasn't seen it
          if (!hasSeenIntro()) {
            throw redirect('/intro');
          }
          return null;
        },
      },
      {
        path: 'gallery',
        element: <GalleryPage />
      },
      {
        path: 'overview',
        element: <TablePage />
      },
      {
        path: 'screenshot/:id',
        element: <ScreenshotDetailPage />,
        action: deleteAction
      },
      {
        path: 'ip/:ip',
        element: <IPPage />
      },
      {
        path: 'ips',
        element: <IPsPage />
      },
      {
        path: 'domains',
        element: <DomainsPage />
      },
      {
        path: 'search',
        element: <SearchResultsPage />,
        action: searchAction,
        loader: searchLoader,
      },
    ]
  }
], {
  basename: getBasename()
});

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
);
