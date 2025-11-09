import ky from 'ky';
import { showDangerToast, showWarningToast } from './toasts';

const httpClient = ky.create({
  hooks: {
    beforeRequest: [
      request => {
        // Add CSRF token to the header for POST requests
        if (request.method === 'POST' || request.method === 'PUT' || request.method === 'PATCH' || request.method === 'DELETE') {
          const token = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
          if (token) {
            request.headers.set('X-CSRF-Token', token);
          }
        }
      }
    ],
    afterResponse: [
      async (request, options, response) => {
        const status = response.status;

        // Exclude 422 as it's used for validation errors
        //  are handled individually by the caller
        if (status >= 400 && status !== 422) {
          let errorMessage = null;

          const contentType = response.headers.get('Content-Type');
          if (contentType && contentType.includes('application/json')) {
            // For JSON responses, display error message if available
            try {
              const jsonData = await response.json();
              errorMessage = jsonData?.error || jsonData?.message || null;
            } catch (e) {
              // Ignore JSON parse errors
            }
          }

          if (!errorMessage) {
            // If an error message is not available from a JSON response or not a JSON response, display the status code as the message
            errorMessage = `An error occurred (HTTP ${status})`;
          }

          if (status >= 500) {
            // Display server errors with danger toast
            showDangerToast(errorMessage);
          } else {
            // Display client errors with warning toast
            showWarningToast(errorMessage);
          }
        }

        return response;
      }
    ],
  }
});

export default httpClient;
