import { createRule, listRule,  updateRule } from './service';


const Model = {
  namespace: 'seoModel',
  state: {
    listData: {
      list: [],
      pagination: {},
    },
  },
  effects: {
    *list({ payload }, { call, put }) {
      const response = yield call(listRule, payload);
      yield put({
        type: '_list',
        payload: response.data,
      });
    },

    *create({ payload, callback }, { call, put }) {
      yield call(createRule, payload);
      if (callback) callback();
      const response = yield call(listRule, payload);
      yield put({
        type: '_list',
        payload: response.data,
      });
    },

    *update({ payload, callback }, { call, put }) {
      yield call(updateRule, payload);
      if (callback) callback();
      const response = yield call(listRule, payload);
      yield put({
        type: '_list',
        payload: response.data,
      });
    },
  },
  reducers: {
    _list(state, action) {
      return {
        ...state,
        listData: action.payload
      };
    },
  },
};
export default Model;
