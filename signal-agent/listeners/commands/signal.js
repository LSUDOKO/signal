const helpBlocks = [
  {
    type: 'header',
    text: { type: 'plain_text', text: '🧘 Signal — Calm Slack for Neurodivergent Professionals', emoji: true },
  },
  {
    type: 'section',
    text: { type: 'mrkdwn', text: 'Here are the commands you can use:' },
  },
  { type: 'divider' },
  {
    type: 'section',
    fields: [
      { type: 'mrkdwn', text: '*/signal*\nOpen this help menu' },
      { type: 'mrkdwn', text: '*/translate [message]*\nDecode ambiguous workplace language' },
    ],
  },
  {
    type: 'section',
    fields: [
      { type: 'mrkdwn', text: '*/catchup [topic]*\nGet AI summary of what you missed' },
      { type: 'mrkdwn', text: '*/focus [duration]*\nStart deep work mode (e.g., /focus 2h)' },
    ],
  },
  {
    type: 'section',
    text: { type: 'mrkdwn', text: '*/digest*\nSend an instant digest' },
  },
];

const signalCommandCallback = async ({ ack, client, command, logger }) => {
  try {
    // Acknowledge immediately to satisfy Slack's 3-second requirement
    await ack();

    // Open DM with the user
    const result = await client.conversations.open({ users: command.user_id });
    const dmChannel = result.channel.id;

    // Post help blocks
    await client.chat.postMessage({
      channel: dmChannel,
      blocks: helpBlocks,
      text: 'Signal Help',
    });
  } catch (error) {
    logger.error('Error handling /signal command:', error);
  }
};

export { signalCommandCallback };
