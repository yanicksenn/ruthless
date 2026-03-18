import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Plus, Library as LibraryIcon, ChevronRight, Hash, LogOut } from 'lucide-react';
import { deckClient, createOptions } from '../api/client';
import { Deck } from '../api/ruthless';
import { CreationDialog } from './CreationDialog';
import { DeckEditor } from './DeckEditor';

export const Decks: React.FC = () => {
  const { token, user, logout, limits } = useAuth();
  const [decks, setDecks] = useState<Deck[]>([]);
  const [loading, setLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [selectedDeckId, setSelectedDeckId] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const decksRes = await deckClient.listDecks({}, createOptions(token));
      setDecks(decksRes.response.decks || []);
    } catch (err) {
      console.error('Failed to fetch decks:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [token]);

  const handleOpenDeckDialog = () => {
    setIsDialogOpen(true);
  };

  const handleCreate = async (value: string) => {
    try {
      await deckClient.createDeck({ name: value }, createOptions(token));
      fetchData();
    } catch (err) {
      alert(`Failed to create deck`);
    }
  };

  if (selectedDeckId) {
    return <DeckEditor deckId={selectedDeckId} onBack={() => setSelectedDeckId(null)} />;
  }

  return (
    <div className="max-w-6xl mx-auto p-4 py-12">
      <header className="flex justify-between items-start mb-12">
        <div>
          <h1 className="text-5xl font-black tracking-tighter text-white">DECKS</h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">
            Manage your custom collections
          </p>
          {user && (
            <div className="mt-4 flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary text-xs font-bold ring-1 ring-primary/30">
                {user.name.slice(0, 2).toUpperCase()}
              </div>
              <span className="text-gray-300 font-bold text-sm tracking-tight">{user.name}</span>
            </div>
          )}
        </div>
        <div className="flex gap-3">
          <button
            onClick={handleOpenDeckDialog}
            className="bg-secondary hover:bg-secondary/80 text-black font-black px-6 py-3 rounded-2xl flex items-center gap-2 transition-all transform hover:scale-105 shadow-lg shadow-secondary/10"
          >
            <Plus size={20} />
            NEW DECK
          </button>
          <button
            onClick={logout}
            className="p-3 bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white rounded-2xl transition-all ring-1 ring-white/5"
            title="Logout"
          >
            <LogOut size={20} />
          </button>
        </div>
      </header>

      <div className="glass p-8 rounded-[2.5rem] min-h-[500px] border border-white/5 shadow-2xl relative overflow-hidden">
        {loading && (
           <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
           </div>
        )}

        <div className="space-y-8">
          <div>
            <h2 className="text-3xl font-black text-white tracking-tight">Secret Stashes</h2>
            <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">Your collections of misery</p>
          </div>
          
          {decks.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-center">
              <div className="p-4 bg-white/5 rounded-full mb-4">
                <LibraryIcon size={48} className="text-gray-700" />
              </div>
              <p className="text-gray-500 font-bold italic mb-6">No decks created yet. Start by creating one!</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {decks.map(deck => (
                <div key={deck.id} onClick={() => setSelectedDeckId(deck.id)} className="group relative glass-light p-6 rounded-3xl border border-white/10 hover:border-primary/30 transition-all cursor-pointer">
                  <div className="flex justify-between items-start mb-4">
                    <div className="p-2 bg-primary/10 rounded-xl">
                      <Hash size={16} className="text-primary" />
                    </div>
                    <div className="px-2 py-1 bg-white/5 rounded-lg text-[10px] font-black tracking-widest text-gray-400">
                      {deck.cardIds.length} CARDS
                    </div>
                  </div>
                  <h3 className="text-xl font-bold text-white mb-6 group-hover:text-primary transition-colors">{deck.name}</h3>
                  <div className="flex justify-end opacity-0 group-hover:opacity-100 transition-opacity">
                     <button className="text-gray-400 hover:text-white p-2">
                        <ChevronRight size={20} />
                     </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
      <CreationDialog
        isOpen={isDialogOpen}
        onClose={() => setIsDialogOpen(false)}
        onCreate={handleCreate}
        title="Forge New Deck"
        placeholder="Enter a name for your collection..."
        label="Deck Name"
        submitLabel="Create Deck"
        maxLength={limits?.maxDeckNameLength}
      />
    </div>
  );
};
