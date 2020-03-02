import request from '@/utils/request';

export async function listRule(params) {
  return request('/admin/tool/list', {
    params,
  });
}

export async function createRule(params) {
  return request('/admin/tool/create', {
    method: 'POST',
    data: { ...params },
  });
}
export async function updateRule(params) {
  return request('/admin/tool/update', {
    method: 'POST',
    data: { ...params },
  });
}
