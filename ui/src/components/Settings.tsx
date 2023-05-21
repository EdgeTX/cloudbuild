import { SettingOutlined } from "@ant-design/icons";
import { Dropdown, MenuProps, Modal } from "antd";
import { useContext, useState } from "react";
import { useForm } from "antd/es/form/Form";
import TokenForm, { FormParameters } from "@comps/TokenForm";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";

function Settings() {
  const [tokenSettingOpened, setTokenSettingOpened] = useState(false);
  const [form] = useForm<FormParameters>();
  const { checkAuth } = useContext(AuthContext) as AuthContextType;

  const onTokenSubmit = (token: string) => {
    localStorage.setItem("token", token);
    checkAuth();
    setTokenSettingOpened(false);
  };

  const onTokenSettingCancel = () => {
    form.resetFields();
    setTokenSettingOpened(false);
  };

  const items: MenuProps["items"] = [
    {
      label: "Access token",
      key: "accessToken",
      onClick: () => setTokenSettingOpened(true),
    },
  ];

  return (
    <>
      <Dropdown menu={{ items }} trigger={["click"]}>
        <SettingOutlined
          style={{ color: "white", fontSize: "1rem" }}
        >
        </SettingOutlined>
      </Dropdown>
      <Modal
        title="Access token setting"
        open={tokenSettingOpened}
        onOk={() => {
          form.submit();
        }}
        onCancel={onTokenSettingCancel}
      >
        <TokenForm form={form} onTokenSubmit={onTokenSubmit} />
      </Modal>
    </>
  );
}

export default Settings;
