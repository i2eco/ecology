import { Button, DatePicker, Form, Input, Modal, Radio, Select, Steps } from 'antd';
import React, { Component } from 'react';

const FormItem = Form.Item;
const { Option } = Select;
const RadioGroup = Radio.Group;

class UpdateForm extends Component {
  static defaultProps = {
    handleUpdate: () => {},
    handleUpdateModalVisible: () => {},
    values: {},
  };

  formLayout = {
    labelCol: {
      span: 7,
    },
    wrapperCol: {
      span: 13,
    },
  };

  constructor(props) {
    super(props);
    this.state = {
      formVals:{
        ...props.values
      }
    };
  }

  okHandle = () => {
    const { form,handleUpdate } = this.props;
    const {formVals} = this.state
    form.validateFields((err, fieldsValue) => {
      if (err) return;
      form.resetFields();
      console.log("formvals",formVals)
      const mergeValue = { ...formVals, ...fieldsValue };
      handleUpdate(mergeValue);
    });
  };


  render() {
    const { updateModalVisible, handleUpdateModalVisible, values,form } = this.props;
    const { formVals } = this.state;
    return (
      <Modal
        width={640}
        bodyStyle={{
          padding: '32px 40px 48px',
        }}
        destroyOnClose
        title="修改猫舍信息"
        visible={updateModalVisible}
        onOk={this.okHandle}
        onCancel={() => handleUpdateModalVisible(false, values)}
        afterClose={() => handleUpdateModalVisible()}
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
            initialValue: formVals.page,
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
            initialValue: formVals.statement,
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
            initialValue: formVals.title,
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
            initialValue: formVals.keywords,
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
            initialValue: formVals.description,
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
  }
}

export default Form.create()(UpdateForm);
