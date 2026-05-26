import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { api } from '@/lib/api';

type WidgetConfig = {
  brand_color: string;
  greeting: string;
  bot_name: string;
  position: string;
  widget_api_key: string;
  is_active: boolean;
};

type WidgetConfigContextType = {
  config: WidgetConfig | null;
  loading: boolean;
  saving: boolean;
  setConfig: React.Dispatch<React.SetStateAction<WidgetConfig | null>>;
  refresh: () => Promise<void>;
  save: () => Promise<void>;
};

const WidgetConfigContext = createContext<WidgetConfigContextType | undefined>(undefined);

export const WidgetConfigProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [config, setConfig] = useState<WidgetConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.get<WidgetConfig>('/widget/config');
      const defaultConfig: WidgetConfig = {
        brand_color: '#3b82f6',
        greeting: 'Hello! How can I help you today? 👋',
        bot_name: 'Noant AI',
        position: 'bottom-right',
        widget_api_key: '',
        is_active: true,
      };
      setConfig({ ...defaultConfig, ...res });
    } catch {
      // keep defaults if fetch fails
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const save = async () => {
    if (!config) return;
    setSaving(true);
    try {
      await api.post('/widget/config', config);
      if (config.is_active) {
        await api.post('/integrations/connect', {
          channel: 'web',
          config: {
            botName: config.bot_name,
            greeting: config.greeting,
            brandColor: config.brand_color,
            position: config.position === 'bottom-left' ? 'left' : 'right',
          }
        });
      } else {
        await api.post('/integrations/disconnect/web');
      }
    } catch (err) {
      console.error('Failed to sync widget config to integrations:', err);
    } finally {
      setSaving(false);
    }
  };

  const refresh = async () => {
    await load();
  };

  return (
    <WidgetConfigContext.Provider value={{ config, loading, saving, setConfig, refresh, save }}>
      {children}
    </WidgetConfigContext.Provider>
  );
};

export const useWidgetConfig = () => {
  const ctx = useContext(WidgetConfigContext);
  if (!ctx) throw new Error('useWidgetConfig must be used within WidgetConfigProvider');
  return ctx;
};
