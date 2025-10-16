import './app.css';
import 'htmx.org';
import Alpine from 'alpinejs';
import NProgress from 'nprogress';
import Toastify from 'toastify-js';

// Load scripts and styles from HTML templates
import 'virtual:template-extract:../views/**/*.html';

Alpine.start();

// -------------------------------------------------------------------------
// display a progress bar for htmx requests
// -------------------------------------------------------------------------
NProgress.configure({ showSpinner: false });
document.addEventListener('htmx:beforeSend', function (e) {
  if (NProgress.isStarted()) return;
  NProgress.start();
});

document.addEventListener('htmx:afterRequest', function (e) {
  // if there is a redirect, keep showing the progress bar until the redirect is complete
  if (e.detail.xhr.getResponseHeader('HX-Location')) return;
  NProgress.done();
});

// -------------------------------------------------------------------------
// error handling
// -------------------------------------------------------------------------
document.addEventListener('htmx:sendError', function (e) {
  // If there is a network error, show a toast notification
  Toastify({
    text: 'A network error has occurred',
    duration: -1,
    newWindow: true,
    close: true,
    gravity: 'top',
    position: 'right',
    stopOnFocus: true,
    style: {
      background: '#f39c12',
    },
  }).showToast();
});

// Swap even for non-200 responses
// Examples:
// 1. To display a 404 page when hx-boost is used
// 2. To redisplay the form for validation errors, which are returned with a 422 status
//    If validation errors are returned with a 200 status, the browser may treat the response as "successful" and prompt to save passwords, etc.
//    Therefore, we return validation errors with a 422 status.
document.addEventListener('htmx:beforeSwap', function (e) {
  const status = e.detail.xhr.status;
  if (status !== 200) {
    e.detail.shouldSwap = true;
  }
});

document.addEventListener('htmx:historyCacheMissLoad', function (e) {
  // When HX-History-Restore-Request is requested, if the response has Hx-Redirect or Hx-Location headers,
  // I want to perform a redirect, but HTMX does not do this.
  // This is a workaround for the current HTMX behavior where it does not interpret these headers on browser back navigation, causing redirects to not work properly.
  const xhr = e.detail.xhr;
  const redirectUrl = xhr.getResponseHeader('HX-Redirect') || xhr.getResponseHeader('HX-Location');

  if (redirectUrl) {
    // Perform the redirect
    window.location.href = redirectUrl;
    e.preventDefault();
  }
});
