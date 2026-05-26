(function () {
  'use strict';

  // ===== CONFIG DEFAULTS =====
  var CONFIG = {
    brandColor: '#0ea5e9',
    greeting: 'Hi! 👋 How can I help you?',
    botName: 'Noant AI',
    position: 'right',
    apiKey: '',
  };

  // ===== DOM HELPERS =====
  function createElement(tag, attrs, children) {
    var el = document.createElement(tag);
    if (attrs) {
      Object.keys(attrs).forEach(function (key) {
        if (key === 'className') {
          el.className = attrs[key];
        } else if (key === 'style') {
          el.style.cssText = attrs[key];
        } else if (key === 'htmlFor') {
          el.setAttribute('for', attrs[key]);
        } else if (key.startsWith('on')) {
          el.addEventListener(key.slice(2).toLowerCase(), attrs[key]);
        } else {
          el.setAttribute(key, attrs[key]);
        }
      });
    }
    if (children) {
      children.forEach(function (child) {
        if (typeof child === 'string') {
          el.appendChild(document.createTextNode(child));
        } else if (child) {
          el.appendChild(child);
        }
      });
    }
    return el;
  }

  function css(selector, rules) {
    var style = document.createElement('style');
    style.textContent = selector + ' {' + rules + '}';
    document.head.appendChild(style);
  }

  // ===== STYLES =====
  css('.noant-widget-btn', [
    'position:fixed',
    'bottom:20px',
    CONFIG.position + ':20px',
    'width:56px',
    'height:56px',
    'border-radius:50%',
    'background:' + CONFIG.brandColor,
    'color:#fff',
    'border:none',
    'cursor:pointer',
    'box-shadow:0 4px 20px rgba(0,0,0,0.2)',
    'z-index:999999',
    'display:flex',
    'align-items:center',
    'justify-content:center',
    'transition:transform 0.2s, box-shadow 0.2s',
    'font-size:28px',
    'line-height:1',
  ].join(';'));

  css('.noant-widget-btn:hover', [
    'transform:scale(1.05)',
    'box-shadow:0 6px 24px rgba(0,0,0,0.25)',
  ].join(';'));

  css('.noant-widget-box', [
    'position:fixed',
    'bottom:90px',
    CONFIG.position + ':20px',
    'width:360px',
    'max-width:calc(100vw - 40px)',
    'height:520px',
    'max-height:calc(100vh - 120px)',
    'border-radius:16px',
    'background:#fff',
    'box-shadow:0 8px 40px rgba(0,0,0,0.18)',
    'z-index:999998',
    'display:flex',
    'flex-direction:column',
    'overflow:hidden',
    'transition:opacity 0.3s, transform 0.3s',
    'opacity:0',
    'transform:translateY(12px)',
    'pointer-events:none',
    'font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif',
  ].join(';'));

  css('.noant-widget-box.open', 'opacity:1;transform:translateY(0);pointer-events:auto;');

  css('.noant-widget-header', [
    'padding:16px 20px',
    'background:' + CONFIG.brandColor,
    'color:#fff',
    'font-size:15px',
    'font-weight:600',
    'display:flex',
    'align-items:center',
    'justify-content:space-between',
    'flex-shrink:0',
  ].join(';'));

  css('.noant-widget-close', [
    'background:none',
    'border:none',
    'color:#fff',
    'font-size:20px',
    'cursor:pointer',
    'padding:0',
    'line-height:1',
    'opacity:0.8',
  ].join(';') + ':hover{opacity:1}');

  css('.noant-widget-messages', [
    'flex:1',
    'overflow-y:auto',
    'padding:16px',
    'display:flex',
    'flex-direction:column',
    'gap:8px',
    'background:#f8fafc',
  ].join(';'));

  css('.noant-msg-ai,.noant-msg-user', [
    'max-width:85%',
    'padding:10px 14px',
    'border-radius:14px',
    'font-size:14px',
    'line-height:1.4',
    'word-wrap:break-word',
  ].join(';'));

  css('.noant-msg-ai', [
    'align-self:flex-start',
    'background:#fff',
    'color:#1e293b',
    'border:1px solid #e2e8f0',
    'border-bottom-left-radius:4px',
  ].join(';'));

  css('.noant-msg-user', [
    'align-self:flex-end',
    'background:' + CONFIG.brandColor,
    'color:#fff',
    'border-bottom-right-radius:4px',
  ].join(';'));

  css('.noant-widget-input-area', [
    'padding:12px 16px',
    'border-top:1px solid #e2e8f0',
    'display:flex',
    'gap:8px',
    'flex-shrink:0',
    'background:#fff',
  ].join(';'));

  css('.noant-widget-input', [
    'flex:1',
    'border:1px solid #cbd5e1',
    'border-radius:10px',
    'padding:10px 14px',
    'font-size:14px',
    'outline:none',
    'font-family:inherit',
  ].join(';') + ':focus{border-color:' + CONFIG.brandColor + '}');

  css('.noant-widget-send', [
    'background:' + CONFIG.brandColor,
    'color:#fff',
    'border:none',
    'border-radius:10px',
    'padding:' + '0 16px',
    'font-size:16px',
    'cursor:pointer',
    'font-weight:600',
    'transition:opacity 0.2s',
  ].join(';') + ':hover{opacity:0.9}:disabled{opacity:0.5;cursor:default}');

  css('.noant-typing', [
    'display:flex',
    'gap:4px',
    'align-items:center',
    'padding:8px 14px',
    'background:#fff',
    'border:1px solid #e2e8f0',
    'border-radius:14px',
    'border-bottom-left-radius:4px',
    'align-self:flex-start',
  ].join(';'));

  css('.noant-typing-dot', [
    'width:8px',
    'height:8px',
    'border-radius:50%',
    'background:' + CONFIG.brandColor,
    'animation:noantBounce 1.4s infinite ease-in-out',
  ].join(';'));

  css('.noant-typing-dot:nth-child(2)', 'animation-delay:0.2s');
  css('.noant-typing-dot:nth-child(3)', 'animation-delay:0.4s');

  css('@keyframes noantBounce', [
    '0%,80%,100%{transform:scale(0.6);opacity:0.4}',
    '40%{transform:scale(1);opacity:1}',
  ].join(''));

  // ===== STATE =====
  var state = {
    open: false,
    messages: [],
    conversationId: null,
    sending: false,
    typing: false,
  };

  // ===== API =====
  function getBaseUrl() {
    // Try to determine the backend URL from where the script is hosted
    var scripts = document.getElementsByTagName('script');
    for (var i = 0; i < scripts.length; i++) {
      var src = scripts[i].src;
      if (src.indexOf('widget.js') !== -1) {
        // Extract base URL from script source
        var base = src.substring(0, src.lastIndexOf('/'));
        // If hosted on the NOANT platform, use API URL
        if (base.indexOf('widget/noant') !== -1 || base.indexOf('noant') !== -1) {
          return base.replace('/widget', '') + '/api/v1';
        }
      }
    }
    return '';
  }

  function sendMessage(text) {
    if (state.sending || !text.trim()) return;
    state.sending = true;
    state.messages.push({ role: 'user', content: text });
    state.typing = true;
    render();

    var baseUrl = getBaseUrl();
    var payload = {
      api_key: CONFIG.apiKey,
      message: text,
    };
    if (state.conversationId) {
      payload.conversation_id = state.conversationId;
    }

    var xhr = new XMLHttpRequest();
    xhr.open('POST', baseUrl + '/widget/public/chat', true);
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.onload = function () {
      state.sending = false;
      state.typing = false;
      if (xhr.status === 200) {
        try {
          var resp = JSON.parse(xhr.responseText);
          if (resp.reply) {
            state.messages.push({ role: 'ai', content: resp.reply });
          }
          if (resp.conversation_id) {
            state.conversationId = resp.conversation_id;
          }
        } catch (e) {
          state.messages.push({ role: 'ai', content: 'Sorry, I had trouble understanding that. Please try again.' });
        }
      } else {
        state.messages.push({ role: 'ai', content: 'Sorry, I\'m having a temporary issue. Please try again later.' });
      }
      render();
    };
    xhr.onerror = function () {
      state.sending = false;
      state.typing = false;
      state.messages.push({ role: 'ai', content: 'Connection error. Please check your internet and try again.' });
      render();
    };
    xhr.send(JSON.stringify(payload));
  }

  // ===== RENDER =====
  var boxEl, btnEl, messagesEl, inputEl, sendEl;

  function render() {
    if (!boxEl || !messagesEl) return;

    // Render messages
    messagesEl.innerHTML = '';
    state.messages.forEach(function (msg) {
      var el = createElement('div', {
        className: msg.role === 'user' ? 'noant-msg-user' : 'noant-msg-ai',
      }, [msg.content]);
      messagesEl.appendChild(el);
    });

    // Show typing indicator
    if (state.typing) {
      var typingEl = createElement('div', { className: 'noant-typing' }, [
        createElement('span', { className: 'noant-typing-dot' }),
        createElement('span', { className: 'noant-typing-dot' }),
        createElement('span', { className: 'noant-typing-dot' }),
      ]);
      messagesEl.appendChild(typingEl);
    }

    // Scroll to bottom
    messagesEl.scrollTop = messagesEl.scrollHeight;

    // Update send button
    if (sendEl) {
      sendEl.disabled = state.sending;
    }
  }

  function buildWidget() {
    // Button
    btnEl = createElement('button', {
      className: 'noant-widget-btn',
      id: 'noant-widget-btn',
      onClick: toggle,
      'aria-label': 'Open chat',
    }, ['💬']);
    document.body.appendChild(btnEl);

    // Chat box
    boxEl = createElement('div', {
      className: 'noant-widget-box',
      id: 'noant-widget-box',
    });

    // Header
    var headerEl = createElement('div', { className: 'noant-widget-header' }, [
      createElement('span', {}, [CONFIG.botName]),
      createElement('button', {
        className: 'noant-widget-close',
        onClick: toggle,
        'aria-label': 'Close chat',
      }, ['✕']),
    ]);
    boxEl.appendChild(headerEl);

    // Messages area
    messagesEl = createElement('div', { className: 'noant-widget-messages', id: 'noant-widget-messages' });

    // Add greeting
    var greetingEl = createElement('div', { className: 'noant-msg-ai' }, [CONFIG.greeting]);
    messagesEl.appendChild(greetingEl);
    state.messages.push({ role: 'ai', content: CONFIG.greeting });

    boxEl.appendChild(messagesEl);

    // Input area
    var inputArea = createElement('div', { className: 'noant-widget-input-area' });
    inputEl = createElement('input', {
      className: 'noant-widget-input',
      type: 'text',
      placeholder: 'Type your message...',
      onKeydown: function (e) {
        if (e.key === 'Enter') {
          handleSend();
        }
      },
    });
    sendEl = createElement('button', {
      className: 'noant-widget-send',
      onClick: handleSend,
    }, ['Send']);
    inputArea.appendChild(inputEl);
    inputArea.appendChild(sendEl);
    boxEl.appendChild(inputArea);

    document.body.appendChild(boxEl);
  }

  function toggle() {
    state.open = !state.open;
    if (state.open) {
      boxEl.classList.add('open');
      btnEl.textContent = '✕';
      setTimeout(function () { inputEl && inputEl.focus(); }, 300);
    } else {
      boxEl.classList.remove('open');
      btnEl.textContent = '💬';
    }
  }

  function handleSend() {
    var text = inputEl.value.trim();
    if (text) {
      inputEl.value = '';
      sendMessage(text);
    }
  }

  // ===== INIT =====
  function init(opts) {
    if (opts) {
      if (opts.brandColor) CONFIG.brandColor = opts.brandColor;
      if (opts.greeting) CONFIG.greeting = opts.greeting;
      if (opts.botName) CONFIG.botName = opts.botName;
      if (opts.position) CONFIG.position = opts.position;
      if (opts.apiKey) CONFIG.apiKey = opts.apiKey;
    }

    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', buildWidget);
    } else {
      buildWidget();
    }
  }

  // Expose global init
  window.NoantWidget = {
    init: init,
    open: function () { if (!state.open) toggle(); },
    close: function () { if (state.open) toggle(); },
  };

  // Auto-init if noantWidgetConfig exists
  if (window.noantWidgetConfig) {
    init(window.noantWidgetConfig);
  }
})();