const GO_BACKEND_URL = process.env.SIGNAL_API_URL || 'http://localhost:8080';

const catchupCommandCallback = async ({ ack, client, command, logger }) => {
  try {
    await ack();

    const result = await client.conversations.open({ users: command.user_id });
    const dmChannel = result.channel.id;

    await client.chat.postMessage({
      channel: dmChannel,
      text: '⏳ Searching Slack and summarizing what you missed...',
    });

    const topic = command.text || 'this week';

    // Search Slack for recent messages about the topic
    const searchQuery = topic === 'this week' ? 'after:this_week' : `${topic} after:this_week`;
    const searchResult = await client.search.messages({
      query: searchQuery,
      sort: 'timestamp',
      sort_dir: 'desc',
      count: 10,
    });

    const messages = searchResult.messages?.matches || [];
    if (messages.length === 0) {
      await client.chat.postMessage({
        channel: dmChannel,
        text: `📭 *Catch-Up: ${topic}*\n\nNo messages found about "${topic}" this week. Try a different topic or check back later.`,
        mrkdwn: true,
      });
      return;
    }

    // Build a summary from search results
    const messageList = messages
      .slice(0, 5)
      .map((m) => `• <#${m.channel.id}>: ${m.text}`)
      .join('\n');

    // Call Go backend AI for a summary
    const aiResponse = await fetch(`${GO_BACKEND_URL}/api/v1/ai/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        system_prompt:
          'You are a "What You Missed" assistant for a neurodivergent professional. Summarize these messages into topics. Highlight decisions, action items, and anything requiring the user\'s input. Be concise and direct.',
        user_prompt: `Summarize what I missed about "${topic}":\n\n${messageList}`,
        max_tokens: 1000,
      }),
    });

    const data = await aiResponse.json();
    const summary = data.text || 'No summary available.';

    await client.chat.postMessage({
      channel: dmChannel,
      text: `📋 *Catch-Up: ${topic}*\n\n${summary}\n\n${messages.length > 5 ? `_Found ${messages.length} messages total — showing the top 5._\n` : ''}_Use \`/catchup "topic"\` anytime._`,
      mrkdwn: true,
    });
  } catch (error) {
    logger.error('Error handling /catchup command:', error);
    try {
      const result = await client.conversations.open({ users: command.user_id });
      await client.chat.postMessage({
        channel: result.channel.id,
        text: '❌ Sorry, I encountered an error. Please try again.',
      });
    } catch (_) {}
  }
};

export { catchupCommandCallback };
