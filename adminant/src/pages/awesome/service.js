import request from '@/utils/request';

export async function listRule(params) {
  return request('/admin/awesome/list', {
    params,
  });
}

export async function createRule(params) {
  return request('/admin/awesome/create', {
    method: 'POST',
    data: { ...params },
  });
}
export async function updateRule(params) {
  return request('/admin/awesome/update', {
    method: 'POST',
    data: { ...params },
  });
}

export async function deleteRule(params) {
  return request('/admin/awesome/delete', {
    method: 'POST',
    data: { ...params },
  });
}

export async function uploadRule(params) {
  return request('/admin/awesome/upload', {
    method: 'POST',
    data: { ...params},
  });
}
