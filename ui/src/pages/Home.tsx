import { Typography } from "antd";
import { useContext } from "react";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";

function Home() {
  const { token } = useContext(AuthContext) as AuthContextType;

  return (
    <>
      <Typography>
        <Typography.Title>Home Page</Typography.Title>
        <Typography.Text>
          Your token is {token}.
        </Typography.Text>
      </Typography>
    </>
  );
}

export default Home;
