import { Form, Input, Modal } from 'antd';
import React from 'react';

const FormItem = Form.Item;

const CreateForm = props => {
  const { modalVisible, form, handleCreate, handleModalVisible } = props;

  const okHandle = () => {
    form.validateFields((err, fieldsValue) => {
      if (err) return;
      form.resetFields();
      handleCreate(fieldsValue);
    });
  };

  return (
    <Modal
      destroyOnClose
      title="新建SEO"
      visible={modalVisible}
      onOk={okHandle}
      onCancel={() => handleModalVisible()}
    >
      <FormItem
        labelCol={{
          span: 5,
        }}
        wrapperCol={{
          span: 15,
        }}
        label="页面"
      >
        {form.getFieldDecorator('page', {
          rules: [
            {
              required: true,
              message: '请输入至少两个字符的规则描述！',
              min: 1,
            },
          ],
        })(<Input placeholder="请输入" />)}
      </FormItem>
      <FormItem
        labelCol={{
          span: 5,
        }}
        wrapperCol={{
          span: 15,
        }}
        label="页面说明"
      >
        {form.getFieldDecorator('statement', {
          rules: [
            {
              required: false,
              message: '请输入至少两个字符的规则描述！',
              min: 2,
            },
          ],
        })(<Input placeholder="请输入" />)}
      </FormItem>

      <FormItem
        labelCol={{
          span: 5,
        }}
        wrapperCol={{
          span: 15,
        }}
        label="SEO标题"
      >
        {form.getFieldDecorator('title', {
          rules: [
            {
              required: false,
              message: '请输入至少两个字符的规则描述！',
              min: 2,
            },
          ],
        })(<Input.TextArea style={{ maxWidth: '95%' }} autosize={{ minRows: 5, maxRows: 10 }} placeholder="请输入" />)}
      </FormItem>
      <FormItem
        labelCol={{
          span: 5,
        }}
        wrapperCol={{
          span: 15,
        }}
        label="SEO关键字"
      >
        {form.getFieldDecorator('keywords', {
          rules: [
            {
              required: false,
              message: '请输入至少两个字符的规则描述！',
              min: 2,
            },
          ],
        })(<Input.TextArea style={{ maxWidth: '95%' }} autosize={{ minRows: 5, maxRows: 10 }} placeholder="请输入" />)}
      </FormItem>
      <FormItem
        labelCol={{
          span: 5,
        }}
        wrapperCol={{
          span: 15,
        }}
        label="SEO摘要"
      >
        {form.getFieldDecorator('description', {
          rules: [
            {
              required: false,
              message: '请输入至少两个字符的规则描述！',
              min: 2,
            },
          ],
        })(<Input.TextArea style={{ maxWidth: '95%' }} autosize={{ minRows: 5, maxRows: 10 }} placeholder="请输入" />)}
      </FormItem>
    </Modal>
  );
};

export default Form.create()(CreateForm);
