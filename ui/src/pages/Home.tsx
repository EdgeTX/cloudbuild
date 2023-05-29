import Layout from "@comps/Layout";
import { Typography } from "antd";

function Home() {
  return (
    <Layout>
    <Typography>
      <Typography.Title>Home Page</Typography.Title>
      <Typography.Text>
        Welcome to Cloudbuild Dashboard.
      </Typography.Text>
    </Typography>
    </Layout>
  );
}

export default Home;
