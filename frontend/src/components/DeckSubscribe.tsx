import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { deckClient, createOptions } from '../api/client';
import { ArrowLeft } from 'lucide-react';

interface DeckSubscribeProps {
  deckId: string;
}

export const DeckSubscribe: React.FC<DeckSubscribeProps> = ({ deckId }) => {
  const { token } = useAuth();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    const subscribe = async () => {
      try {
        await deckClient.subscribeToDeck({ deckId }, createOptions(token));
        setStatus('success');
        // Redirect to decks after short delay
        setTimeout(() => {
          window.location.href = '/decks';
        }, 1500);
      } catch (err: any) {
        setStatus('error');
        setErrorMsg(err.message || 'Failed to subscribe to deck');
      }
    };

    subscribe();
  }, [deckId, token]);

  if (status === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
          <p className="text-gray-400 font-bold tracking-widest uppercase">Subscribing to deck...</p>
        </div>
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <div className="glass p-8 rounded-[2.5rem] border border-red-500/20 text-center max-w-md w-full">
          <h2 className="text-2xl font-black text-white mb-4">Subscription Failed</h2>
          <p className="text-red-400 mb-8">{errorMsg}</p>
          <button
            onClick={() => window.location.href = '/decks'}
            className="flex items-center justify-center gap-2 w-full bg-white/5 hover:bg-white/10 text-white p-4 rounded-2xl font-black transition-colors"
          >
            <ArrowLeft size={20} />
            BACK TO DECKS
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="glass p-8 rounded-[2.5rem] border border-green-500/20 text-center max-w-md w-full">
        <h2 className="text-3xl font-black text-green-400 mb-4">Success!</h2>
        <p className="text-gray-300 font-bold mb-2">You are now subscribed to the deck.</p>
        <p className="text-gray-500 text-sm mb-8">Redirecting you to your decks...</p>
      </div>
    </div>
  );
};
