import React, { useState, useEffect } from 'react';
import { Decks } from './Decks';
import { Cards } from './Cards';
import { useAuth } from '../context/AuthContext';
import { LogOut } from 'lucide-react';

interface LibraryProps {
  initialTab?: 'decks' | 'cards';
  activeDeckId: string | null;
  activeTab: string | null;
  onSelectDeck: (id: string | null) => void;
}

export const Library: React.FC<LibraryProps> = ({ initialTab = 'decks', activeDeckId, activeTab: deckEditorTab, onSelectDeck }) => {
  const { user, logout } = useAuth();
  const [activeTab, setActiveTab] = useState<'decks' | 'cards'>(initialTab);

  // Sync tab with URL
  useEffect(() => {
    // Only update URL if not in deck editor
    if (!activeDeckId) {
       const newPath = `/library/${activeTab}`;
       if (window.location.pathname !== newPath) {
         window.history.pushState(null, '', newPath);
       }
    }
  }, [activeTab, activeDeckId]);

  // If a deck is selected, we show the Decks component (which will show DeckEditor)
  if (activeDeckId) {
    return <Decks activeDeckId={activeDeckId} activeTab={deckEditorTab} onSelectDeck={onSelectDeck} />;
  }

  return (
    <div className="max-w-6xl mx-auto p-4 py-12">
      <header className="flex justify-between items-start mb-8">
        <div>
          <h1 className="text-5xl font-black tracking-tighter text-white">LIBRARY</h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">
            Manage your collection of decks and cards
          </p>
          {user && (
            <div className="mt-4 flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary text-xs font-bold ring-1 ring-primary/30">
                {user.name.slice(0, 2).toUpperCase()}
              </div>
              <span className="text-gray-300 font-bold text-sm tracking-tight">
                {user.name}
                {user.identifier && <span className="text-gray-500 italic">#{user.identifier}</span>}
              </span>
            </div>
          )}
        </div>
        <button
          onClick={logout}
          className="p-3 bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white rounded-2xl transition-all ring-1 ring-white/5"
          title="Logout"
        >
          <LogOut size={20} />
        </button>
      </header>

      {/* Sub-menu - Same style as DeckEditor */}
      <div className="flex gap-4 mb-8 border-b border-white/10 pb-4">
        <button
          onClick={() => setActiveTab('decks')}
          className={`px-6 py-2 rounded-xl font-black text-sm transition-all tracking-widest uppercase ${
            activeTab === 'decks' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white hover:bg-white/5'
          }`}
        >
          Decks
        </button>
        <button
          onClick={() => setActiveTab('cards')}
          className={`px-6 py-2 rounded-xl font-black text-sm transition-all tracking-widest uppercase ${
            activeTab === 'cards' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white hover:bg-white/5'
          }`}
        >
          Cards
        </button>
      </div>

      <div>
        {activeTab === 'decks' ? (
          <Decks activeDeckId={activeDeckId} activeTab={deckEditorTab} onSelectDeck={onSelectDeck} />
        ) : (
          <Cards />
        )}
      </div>
    </div>
  );
};
