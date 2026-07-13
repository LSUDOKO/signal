#!/bin/bash

echo "=========================================="
echo "Signal Feature Testing Script"
echo "=========================================="
echo ""
echo "This script will guide you through testing all features."
echo "Make sure your server is running: go run ./cmd/api/main.go"
echo ""
read -p "Press Enter to continue..."
echo ""

echo "✅ TEST 1: @Signal Mention"
echo "----------------------------------------"
echo "In any Slack channel, type:"
echo ""
echo "    @Signal hi"
echo ""
echo "Expected: Signal responds with welcome message"
read -p "Did it work? (y/n): " test1
echo ""

echo "✅ TEST 2: Social Translator"
echo "----------------------------------------"
echo "In any Slack channel or DM, type:"
echo ""
echo "    /translate per my last email, we need this by EOD"
echo ""
echo "Expected: Signal DMs you with tone analysis"
read -p "Did it work? (y/n): " test2
echo ""

echo "✅ TEST 3: Catch-Up (Semantic Search)"
echo "----------------------------------------"
echo "In any Slack channel or DM, type:"
echo ""
echo "    /catchup what's happening with the project"
echo ""
echo "Expected: Signal DMs you with AI summary"
read -p "Did it work? (y/n): " test3
echo ""

echo "✅ TEST 4: Help Menu"
echo "----------------------------------------"
echo "In any Slack channel or DM, type:"
echo ""
echo "    /signal"
echo ""
echo "Expected: Signal DMs you with command list"
read -p "Did it work? (y/n): " test4
echo ""

echo "=========================================="
echo "Test Results Summary"
echo "=========================================="
echo ""

passed=0
total=4

if [ "$test1" = "y" ]; then
  echo "✅ @Signal Mention: PASSED"
  ((passed++))
else
  echo "❌ @Signal Mention: FAILED"
fi

if [ "$test2" = "y" ]; then
  echo "✅ Social Translator: PASSED"
  ((passed++))
else
  echo "❌ Social Translator: FAILED"
fi

if [ "$test3" = "y" ]; then
  echo "✅ Catch-Up: PASSED"
  ((passed++))
else
  echo "❌ Catch-Up: FAILED"
fi

if [ "$test4" = "y" ]; then
  echo "✅ Help Menu: PASSED"
  ((passed++))
else
  echo "❌ Help Menu: FAILED"
fi

echo ""
echo "=========================================="
echo "FINAL SCORE: $passed/$total tests passed"
echo "=========================================="

if [ $passed -eq $total ]; then
  echo ""
  echo "🎉 ALL TESTS PASSED! 🎉"
  echo "Your Signal bot is fully functional!"
  echo ""
  echo "Next steps:"
  echo "1. Record demo video (see FINAL_STATUS_AND_TESTING.md)"
  echo "2. Create architecture diagram"
  echo "3. Submit to Devpost"
  echo ""
elif [ $passed -ge 3 ]; then
  echo ""
  echo "⚠️ MOST TESTS PASSED"
  echo "You have enough working features for submission."
  echo "Review failed tests in server logs if needed."
  echo ""
else
  echo ""
  echo "❌ CRITICAL ISSUES"
  echo "Please check server logs and review:"
  echo "  - FINAL_STATUS_AND_TESTING.md"
  echo "  - SLACK_EVENT_SUBSCRIPTIONS_FIX.md"
  echo ""
fi
