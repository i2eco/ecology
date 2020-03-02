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
      title="新建工具"
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
        label="名称"
      >
        {form.getFieldDecorator('name', {
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
        label="描述"
      >
        {form.getFieldDecorator('desc', {
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
        label="唯一标识"
      >
        {form.getFieldDecorator('identify', {
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
