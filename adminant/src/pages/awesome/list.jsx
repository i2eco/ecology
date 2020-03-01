import {
  Button,
  Card,
  Col,
  Form,
  Input,
  Row,
  Select,
  message,
  Upload,
  Popconfirm
} from 'antd';
import React, { Component, Fragment } from 'react';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import { connect } from 'dva';
import moment from 'moment';
import styles from './style.less';
import CreateForm from './components/CreateForm';
import StandardTable from '@/components/SaberStandardTable';
import UpdateForm from './components/UpdateForm';
const UPLOADFile = '/admin/awesome/upload';

const FormItem = Form.Item;
const { Option } = Select;

const getValue = obj =>
  Object.keys(obj)
    .map(key => obj[key])
    .join(',');


/* eslint react/no-multi-comp:0 */
@connect(({ awesomeModel, loading }) => ({
  listData:awesomeModel.listData,
  loading: loading.models.awesomeModel,
}))
class TableList extends Component {
  state = {
    modalVisible: false,
    updateModalVisible: false,
    expandForm: false,
    formValues: {},
    updateInitialValues: {},
  };
  columns = [
    {
      title: '编号',
      dataIndex: 'id',
    },
    {
      title: '名称',
      dataIndex: 'name',
    },
    {
      title: '说明',
      dataIndex: 'desc',
    },
    {
      title: 'Star',
      dataIndex: 'starCount',
    },
    {
      title: 'Git更新时间',
      dataIndex: 'gitUpdatedAt',
      render: val => <span>{moment(val).format('YYYY-MM-DD HH:mm:ss')}</span>,
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      sorter: true,
      render: val => <span>{moment(val).format('YYYY-MM-DD HH:mm:ss')}</span>,
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      sorter: true,
      render: val => <span>{moment(val).format('YYYY-MM-DD HH:mm:ss')}</span>,
    },
    {
      title: '操作',
      render: (text, record) => (
        <Fragment>
          <a onClick={() => this.handleUpdateModalVisible(true, record)}>修改</a>
          <Popconfirm title="是否要删除此行？" onConfirm={() => this.handleDelete(record.id)}>
            <a>删除</a>
          </Popconfirm>
        </Fragment>

      ),
    },
  ];

  componentDidMount() {
    const { dispatch } = this.props;
    dispatch({
      type: 'awesomeModel/list',
    });
  }

  handleStandardTableChange = (pagination, filtersArg, sorter) => {
    const { dispatch } = this.props;
    const { formValues } = this.state;
    const filters = Object.keys(filtersArg).reduce((obj, key) => {
      const newObj = { ...obj };
      newObj[key] = getValue(filtersArg[key]);
      return newObj;
    }, {});
    const params = {
      currentPage: pagination.current,
      pageSize: pagination.pageSize,
      ...formValues,
      ...filters,
    };

    if (sorter.field) {
      params.sorter = `${sorter.field}_${sorter.order}`;
    }

    dispatch({
      type: 'awesomeModel/list',
      payload: params,
    });
  };

  handleFormReset = () => {
    const { form, dispatch } = this.props;
    form.resetFields();
    this.setState({
      formValues: {},
    });
    dispatch({
      type: 'awesomeModel/list',
      payload: {},
    });
  };

  handleSearch = e => {
    e.preventDefault();
    const { dispatch, form } = this.props;
    form.validateFields((err, fieldsValue) => {
      if (err) return;
      const values = {
        ...fieldsValue,
        updatedAt: fieldsValue.updatedAt && fieldsValue.updatedAt.valueOf(),
      };
      this.setState({
        formValues: values,
      });
      dispatch({
        type: 'awesomeModel/list',
        payload: values,
      });
    });
  };


  handleCreate = fields => {
    const { dispatch } = this.props;
    dispatch({
      type: 'awesomeModel/create',
      payload: {
        ...fields,
      },
    });
    message.success('添加成功');
    this.handleModalVisible();
  };

  handleUpdate = fields => {
    const { dispatch } = this.props;
    dispatch({
      type: 'awesomeModel/update',
      payload: {
        ...fields,
      },
    });
    message.success('配置成功');
    this.handleUpdateModalVisible();
  };

  handleDelete = id => {
    const { dispatch } = this.props;
    dispatch({
      type: 'awesomeModel/delete',
      payload: {
        id,
      },
    });
    message.success('成功');
    this.handleModalVisible();
  };

  handleModalVisible = flag => {
    this.setState({
      modalVisible: !!flag,
    });
  };

  handleUpdateModalVisible = (flag, record) => {
    this.setState({
      updateModalVisible: !!flag,
      updateInitialValues: record || {},
    });
  };


  renderSimpleForm() {
    const { form } = this.props;
    const { getFieldDecorator } = form;
    return (
      <Form onSubmit={this.handleSearch} layout="inline">
        <Row
          gutter={{
            md: 8,
            lg: 24,
            xl: 48,
          }}
        >
          <Col md={6} sm={24}>
            <FormItem label="名称">
              {getFieldDecorator('name')(<Input placeholder="请输入名称" />)}
            </FormItem>
          </Col>
          <Col md={6} sm={24}>
            <FormItem label="繁育人">
              {getFieldDecorator('author')(<Input placeholder="请输入繁育人" />)}
            </FormItem>
          </Col>
          <Col md={6} sm={24}>
            <FormItem label="地址">
              {getFieldDecorator('address')(<Input placeholder="请输入地址" />)}
            </FormItem>
          </Col>
          <Col md={6} sm={24}>
            <span className={styles.submitButtons}>
              <Button type="primary" htmlType="submit">
                查询
              </Button>
              <Button
                style={{
                  marginLeft: 8,
                }}
                onClick={this.handleFormReset}
              >
                重置
              </Button>
            </span>
          </Col>
        </Row>
      </Form>
    );
  }

  render() {
    const {
      listData,
      loading,
    } = this.props;
    const { modalVisible, updateModalVisible, updateInitialValues } = this.state;

    return (
      <PageHeaderWrapper>
        <Card bordered={false}>
          <div className={styles.tableList}>
            <div className={styles.tableListForm}>{this.renderSimpleForm()}</div>
            <div className={styles.tableListOperator}>
              <Button icon="plus" type="primary" onClick={() => this.handleModalVisible(true)}>
                新建
              </Button>
              <Upload
                name="file-upload"
                action={UPLOADFile}
                supportServerRender="true"
                showUploadList={false}
                withCredentials="true"
              >
                <Button icon="plus" type="primary">
                  上传
                </Button>
              </Upload>
            </div>
            <StandardTable
              loading={loading}
              data={listData}
              columns={this.columns}
              onChange={this.handleStandardTableChange}
            />
          </div>

        </Card>
        <CreateForm
          handleCreate={this.handleCreate}
          handleModalVisible={this.handleModalVisible}
          modalVisible={modalVisible}
        />
        {updateInitialValues && Object.keys(updateInitialValues).length ? (
          <UpdateForm
            handleUpdateModalVisible={this.handleUpdateModalVisible}
            handleUpdate={this.handleUpdate}
            updateModalVisible={updateModalVisible}
            values={updateInitialValues}
          />
        ) : null}
      </PageHeaderWrapper>
    );
  }
}

export default Form.create()(TableList);
