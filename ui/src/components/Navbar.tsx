import { Col, Menu, MenuProps, Row } from "antd";
import { Header } from "antd/es/layout/layout";
import { Link, useLocation } from "react-router-dom";
import Settings from "@comps/Settings";

const items: MenuProps["items"] = [
  {
    label: <Link to="/home">Home</Link>,
    key: "/home",
  },
  {
    label: <Link to="/jobs">Jobs</Link>,
    key: "/jobs",
  },
  {
    label: <Link to="/create">Create</Link>,
    key: "/create",
  },
  {
    label: <Link to="/workers">Workers</Link>,
    key: "/workers",
  }
];

function Navbar() {
  const location = useLocation();
  return (
    <Header>
      <Row align="middle">
        <Col flex="auto">
          <Menu
            theme="dark"
            mode="horizontal"
            defaultSelectedKeys={[location.pathname]}
            items={items}
          />
        </Col>
        <Col>
          <Settings />
        </Col>
      </Row>
    </Header>
  );
}

export default Navbar;
