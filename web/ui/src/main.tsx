import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import '@/index.css';

import App from '@/pages/App';
import ErrorPage from '@/pages/Error';

import DashboardPage from '@/pages/dashboard/Dashboard';
import GalleryPage from '@/pages/gallery/Gallery';
import TablePage from '@/pages/table/Table';
import ScreenshotDetailPage from '@/pages/detail/Detail';
import SearchResultsPage from '@/pages/search/Search';
import JobSubmissionPage from '@/pages/submit/Submit';
import IPPage from '@/pages/ip/IP';
import SettingsPage from '@/pages/settings/Settings';

import { searchAction } from '@/pages/search/action';
import { searchLoader } from '@/pages/search/loader';
import { deleteAction } from '@/pages/detail/actions';
import { submitImmediateAction, submitJobAction } from '@/pages/submit/action';

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
    path: '/',
    element: <App />,
    errorElement: <ErrorPage />,
    children: [
      {
        path: '/',
        element: <DashboardPage />
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
        path: 'search',
        element: <SearchResultsPage />,
        action: searchAction,
        loader: searchLoader,
      },
      {
        path: 'submit',
        element: <JobSubmissionPage />,
        action: async ({ request }) => {
          const formData = await request.formData();
          const action = formData.get('action');

          switch (action) {
            case 'job':
              return submitJobAction({ formData });
            case 'immediate':
              return submitImmediateAction({ formData });

            default:
              throw new Error('unknown action for job submit route');
          }
        },
      },
      {
        path: 'settings',
        element: <SettingsPage />
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
