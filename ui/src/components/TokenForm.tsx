import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";
import { Form, FormInstance, Input } from "antd";
import { useContext } from "react";

interface FormParameters {
  accessKey: string;
  secretKey: string;
}

interface Props {
  form: FormInstance<FormParameters>;
  onTokenSubmit: (token: string) => void;
}

function TokenForm({ form, onTokenSubmit }: Props) {
  const { token } = useContext(AuthContext) as AuthContextType;
  const [accessKey, secretKey] = token.split("-");

  const onFinish = (values: FormParameters) => {
    const token = `${values.accessKey}-${values.secretKey}`;
    onTokenSubmit(token);
  };

  return (
    <Form
      labelCol={{ span: 5 }}
      form={form}
      name="Login"
      onFinish={onFinish}
      initialValues={{ accessKey, secretKey }}
      style={{ marginTop: "2rem", marginBottom: "2rem" }}
    >
      <Form.Item
        label="Access key"
        name="accessKey"
        rules={[{ required: true, len: 16 }]}
      >
        <Input />
      </Form.Item>
      <Form.Item
        label="Secret key"
        name="secretKey"
        rules={[{ required: true, len: 32 }]}
      >
        <Input.Password />
      </Form.Item>
    </Form>
  );
}

export type { FormParameters };
export default TokenForm;
