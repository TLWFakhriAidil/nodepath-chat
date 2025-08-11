import { useSearchParams } from 'react-router-dom';
import ChatSimulation from '@/components/ChatSimulation';

const TestChat = () => {
  const [searchParams] = useSearchParams();
  const flowId = searchParams.get('flowId');

  return (
    <ChatSimulation preselectedFlowId={flowId} />
  );
};

export default TestChat;