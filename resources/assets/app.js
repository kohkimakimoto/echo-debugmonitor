import 'htmx.org';
import Alpine from 'alpinejs';
import NProgress from 'nprogress';

import 'extract-sfc-script:../views/**/*.html';

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
  // If there is a redirect, keep showing the progress bar until the redirect is complete
  if (e.detail.xhr.getResponseHeader('HX-Location') || e.detail.xhr.getResponseHeader('HX-Redirect')) return;
  NProgress.done();
});
