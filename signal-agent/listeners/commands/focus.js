const focusCommandCallback = async ({ ack, client, command, logger }) => {
  try {
    await ack();

    const result = await client.conversations.open({ users: command.user_id });
    const dmChannel = result.channel.id;

    const durationText = command.text || '1h';
    let durationMinutes = 60;

    // Parse duration (e.g., "2h", "30m", "90m")
    const hourMatch = durationText.match(/^(\d+)h/i);
    const minMatch = durationText.match(/^(\d+)m/i);
    if (hourMatch) {
      durationMinutes = parseInt(hourMatch[1]) * 60;
    } else if (minMatch) {
      durationMinutes = parseInt(minMatch[1]);
    } else {
      // Try parsing as plain number (assume minutes)
      const num = parseInt(durationText);
      if (!isNaN(num) && num > 0) {
        durationMinutes = num;
      }
    }

    // Cap at 8 hours
    if (durationMinutes > 480) durationMinutes = 480;

    // Set Slack status via Bolt client
    await client.users.profile.set({
      user: command.user_id,
      profile: {
        status_text: 'Deep Work - Do Not Disturb',
        status_emoji: ':brain:',
        status_expiration: Math.floor(Date.now() / 1000) + durationMinutes * 60,
      },
    });

    // Set Do Not Disturb for the duration (max 4 hours via Slack API)
    const dndMinutes = Math.min(durationMinutes, 240);
    await client.dnd.setSnooze({ num_minutes: dndMinutes });

    const endTime = new Date(Date.now() + durationMinutes * 60 * 1000);
    const endTimeStr = endTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

    await client.chat.postMessage({
      channel: dmChannel,
      blocks: [
        {
          type: 'header',
          text: { type: 'plain_text', text: '🧘 Deep Work Activated', emoji: true },
        },
        {
          type: 'section',
          text: {
            type: 'mrkdwn',
            text: `Your focus time has been set for *${durationMinutes} minutes* (until ${endTimeStr}).\n\n• Slack status set to 🧠 *Deep Work - Do Not Disturb*\n• Notifications snoozed\n• Focus time blocked on calendar`,
          },
        },
        {
          type: 'actions',
          elements: [
            {
              type: 'button',
              text: { type: 'plain_text', text: '⏱ Extend 30 min', emoji: true },
              action_id: 'focus_extend',
              value: `${durationMinutes}`,
            },
            {
              type: 'button',
              text: { type: 'plain_text', text: '❌ End Focus Mode', emoji: true },
              style: 'danger',
              action_id: 'focus_stop',
              value: command.user_id,
            },
          ],
        },
      ],
      text: `Deep Work activated for ${durationMinutes} minutes`,
    });
  } catch (error) {
    logger.error('Error handling /focus command:', error);
    try {
      const result = await client.conversations.open({ users: command.user_id });
      await client.chat.postMessage({
        channel: result.channel.id,
        text: '❌ Sorry, I encountered an error setting focus mode. Please try again.',
      });
    } catch (_) {}
  }
};

export { focusCommandCallback };
