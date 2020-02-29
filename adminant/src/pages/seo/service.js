import request from '@/utils/request';

export async function listRule(params) {
  return request('/admin/seo/list', {
    params,
  });
}

export async function createRule(params) {
  return request('/admin/seo/create', {
    method: 'POST',
    data: { ...params },
  });
}
export async function updateRule(params) {
  return request('/admin/seo/update', {
    method: 'POST',
    data: { ...params },
  });
}
