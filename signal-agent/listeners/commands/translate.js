const GO_BACKEND_URL = process.env.SIGNAL_API_URL || 'http://localhost:8080';
const AI_MODEL = process.env.OPENAI_MODEL || 'llama-3.3-70b-versatile';

const translateCommandCallback = async ({ ack, client, command, logger }) => {
  try {
    await ack();

    // Open DM with the user
    const result = await client.conversations.open({ users: command.user_id });
    const dmChannel = result.channel.id;

    // Send "Working..." message
    await client.chat.postMessage({
      channel: dmChannel,
      text: '⏳ Translating your message...',
    });

    const messageToTranslate = command.text || '';

    if (!messageToTranslate) {
      await client.chat.postMessage({
        channel: dmChannel,
        text: 'Please provide a message to translate: `/translate "your message here"`',
      });
      return;
    }

    // Call Go backend AI for translation
    const aiResponse = await fetch(`${GO_BACKEND_URL}/api/v1/ai/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        system_prompt:
          'You are a direct, kind translator of workplace subtext for autistic and ADHD adults. You decode ambiguous messages into literal meaning without judgment. Respond in this exact format:\n- Tone: [single word or short phrase]\n- Intent: [1 sentence, what they want]\n- Action: [1 sentence, what you should do]\n- Note: [1-2 sentences explaining any hidden social context]\n\nRules:\n- Never say "they might be" — be confident but not rude.\n- Never use "just" or "simply" — these are patronizing.\n- If the message is genuinely neutral, say so clearly.\n- If the message is passive-aggressive, name it directly but kindly.',
        user_prompt: `Analyze this message: "${messageToTranslate}"`,
        max_tokens: 500,
      }),
    });

    const data = await aiResponse.json();
    const analysis = data.text || 'Could not analyze the message. Please try again.';

    await client.chat.postMessage({
      channel: dmChannel,
      text: `🔍 *Social Translator*\n\nYou said: "${messageToTranslate}"\n\n${analysis}\n\n_Use \`/translate "message"\` anytime to decode workplace language._`,
      mrkdwn: true,
    });
  } catch (error) {
    logger.error('Error handling /translate command:', error);
    try {
      const result = await client.conversations.open({ users: command.user_id });
      await client.chat.postMessage({
        channel: result.channel.id,
        text: '❌ Sorry, I encountered an error processing your translation. Please try again.',
      });
    } catch (_) {
      // ignore
    }
  }
};

export { translateCommandCallback };
