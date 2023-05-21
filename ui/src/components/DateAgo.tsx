import { Tooltip, Typography } from "antd";
const { Text } = Typography;

import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import localizedFormat from "dayjs/plugin/localizedFormat";
dayjs.extend(relativeTime);
dayjs.extend(localizedFormat);

interface Props {
  date: string;
}

function DateAgo({ date }: Props) {
  const parsedDate = dayjs(date);
  return (
    <Tooltip title={parsedDate.format("LLL")}>
      <Text>{parsedDate.fromNow()}</Text>
    </Tooltip>
  );
}

export default DateAgo;
