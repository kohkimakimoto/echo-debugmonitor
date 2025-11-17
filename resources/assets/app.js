import 'htmx.org';
import Alpine from 'alpinejs';
import collapse from '@alpinejs/collapse';
import NProgress from 'nprogress';
import { showDangerToast } from "./toasts";

Alpine.plugin(collapse);
window.Alpine = Alpine;
Alpine.start();

// ----------------------------------------------------------------------------------------------
// Show progress bar for htmx requests
// ----------------------------------------------------------------------------------------------
NProgress.configure({ showSpinner: false });
document.addEventListener('htmx:beforeRequest', function (e) {
  if (NProgress.isStarted()) return;
  NProgress.start();
});

document.addEventListener('htmx:afterRequest', function (e) {
  // Keep showing the progress bar until the redirect is complete
  if (e.detail.xhr.getResponseHeader('HX-Location')) return;
  NProgress.done();
});

// ----------------------------------------------------------------------------------------------
// Handle network errors for htmx requests
// ----------------------------------------------------------------------------------------------
document.addEventListener('htmx:sendError', function (e) {
  // Show a danger toast for network errors since there is no content to display from the server
  showDangerToast('Network error occurred');
});
