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
        title="修改工具信息"
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
          label="名称"
        >
          {form.getFieldDecorator('name', {
            initialValue: formVals.name,
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
          label="说明"
        >
          {form.getFieldDecorator('desc', {
            initialValue: formVals.desc,
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
            initialValue: formVals.identify,
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
