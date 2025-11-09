import Toastify from 'toastify-js';

const presetOptions = {
  danger: {
    duration: -1, // display indefinitely
    style: {
      background: '#e7000b',
    },
  },
  warning: {
    duration: 5000,
    style: {
      background: '#f39c12',
    },
  },
  success: {
    duration: 5000,
    style: {
      background: '#00a65a',
    },
  },
  info: {
    duration: 5000,
    style: {
      background: '#00c0ef',
    },
  },
};

const commonOptions = {
  newWindow: true,
  close: true,
  gravity: 'top',
  position: 'right',
  stopOnFocus: true,
};

export function showDangerToast(text) {
  Toastify({
    text,
    ...commonOptions,
    ...presetOptions.danger,
  }).showToast();
}

export function showWarningToast(text) {
  Toastify({
    text,
    ...commonOptions,
    ...presetOptions.warning,
  }).showToast();
}

export function showSuccessToast(text) {
  Toastify({
    text,
    ...commonOptions,
    ...presetOptions.success,
  }).showToast();
}

export function showInfoToast(text) {
  Toastify({
    text,
    ...commonOptions,
    ...presetOptions.info,
  }).showToast();
}
