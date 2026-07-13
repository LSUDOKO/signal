const GO_BACKEND_URL = process.env.SIGNAL_API_URL || 'http://localhost:8080';

const digestCommandCallback = async ({ ack, client, command, logger }) => {
  try {
    await ack();

    const result = await client.conversations.open({ users: command.user_id });
    const dmChannel = result.channel.id;

    await client.chat.postMessage({
      channel: dmChannel,
      text: '⏳ Gathering your digest...',
    });

    // Search for recent mentions via Slack Search API
    const searchResult = await client.search.messages({
      query: `from:@${command.user_id} OR to:@${command.user_id} after:today`,
      sort: 'timestamp',
      sort_dir: 'desc',
      count: 15,
    });

    const matches = searchResult.messages?.matches || [];

    if (matches.length === 0) {
      await client.chat.postMessage({
        channel: dmChannel,
        blocks: [
          {
            type: 'header',
            text: { type: 'plain_text', text: '📬 No Mentions Today', emoji: true },
          },
          {
            type: 'section',
            text: {
              type: 'mrkdwn',
              text: 'You have no new mentions or direct messages today. Enjoy the quiet! :sparkles:',
            },
          },
        ],
        text: 'No mentions today',
      });
      return;
    }

    // Categorize messages
    const urgent = [];
    const fyi = [];
    const threadReplies = [];

    for (const match of matches.slice(0, 10)) {
      const text = match.text || '';
      const channelName = match.channel?.name || 'unknown';
      const permalink = match.permalink || '';

      // Simple heuristic: mentions with "urgent/asap/deadline" are urgent
      if (/urgent|asap|deadline|today|eod|critical/i.test(text)) {
        urgent.push({ channel: channelName, text, link: permalink });
      } else if (text.includes('@' + command.user_id)) {
        threadReplies.push({ channel: channelName, text, link: permalink });
      } else {
        fyi.push({ channel: channelName, text, link: permalink });
      }
    }

    // Build categorized digest blocks
    const blocks = [
      {
        type: 'header',
        text: { type: 'plain_text', text: '📬 On-Demand Digest', emoji: true },
      },
    ];

    if (urgent.length > 0) {
      const urgentText = urgent
        .map((m) => `• <#${m.channel}>: "${m.text.substring(0, 100)}"`)
        .join('\n');
      blocks.push({
        type: 'section',
        text: {
          type: 'mrkdwn',
          text: `*🔴 Urgent (needs response today)*\n${urgentText}`,
        },
      });
    }

    if (fyi.length > 0) {
      const fyiText = fyi
        .map((m) => `• <#${m.channel}>: "${m.text.substring(0, 100)}"`)
        .join('\n');
      blocks.push({
        type: 'section',
        text: {
          type: 'mrkdwn',
          text: `*🟢 FYI (no action needed)*\n${fyiText}`,
        },
      });
    }

    if (threadReplies.length > 0) {
      const threadText = threadReplies
        .map((m) => `• <#${m.channel}>: "${m.text.substring(0, 100)}"`)
        .join('\n');
      blocks.push({
        type: 'section',
        text: {
          type: 'mrkdwn',
          text: `*💬 Thread Replies*\n${threadText}`,
        },
      });
    }

    blocks.push(
      { type: 'divider' },
      {
        type: 'context',
        elements: [
          {
            type: 'mrkdwn',
            text: `_Found ${matches.length} messages. Use \`/digest\` anytime or set Quiet Hours in preferences for automatic delivery._`,
          },
        ],
      }
    );

    await client.chat.postMessage({
      channel: dmChannel,
      blocks,
      text: 'Your on-demand digest',
    });
  } catch (error) {
    logger.error('Error handling /digest command:', error);
    try {
      const result = await client.conversations.open({ users: command.user_id });
      await client.chat.postMessage({
        channel: result.channel.id,
        text: '❌ Sorry, I encountered an error generating your digest. Please try again.',
      });
    } catch (_) {}
  }
};

export { digestCommandCallback };
