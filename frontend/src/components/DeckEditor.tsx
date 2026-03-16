import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Plus, Trash2, ArrowLeft } from 'lucide-react';
import { cardClient, deckClient, createOptions } from '../api/client';
import { Card, Deck } from '../api/ruthless';

interface DeckEditorProps {
  deckId: string;
  onBack: () => void;
}

export const DeckEditor: React.FC<DeckEditorProps> = ({ deckId, onBack }) => {
  const { token, user } = useAuth();
  const [deck, setDeck] = useState<Deck | null>(null);
  const [deckCards, setDeckCards] = useState<Card[]>([]);
  const [availableCards, setAvailableCards] = useState<Card[]>([]);
  
  const [loadingDeck, setLoadingDeck] = useState(true);
  const [loadingAvailable, setLoadingAvailable] = useState(true);
  
  const [pageNumber, setPageNumber] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const pageSize = 12;

  const fetchDeckAndCards = async () => {
    setLoadingDeck(true);
    try {
      const deckRes = await deckClient.getDeck({ id: deckId }, createOptions(token));
      const currentDeck = deckRes.response;
      setDeck(currentDeck);

      if (currentDeck.cardIds && currentDeck.cardIds.length > 0) {
        const cardsRes = await cardClient.listCards({ 
          pageSize: 0, 
          pageNumber: 0, 
          ids: currentDeck.cardIds 
        }, createOptions(token));
        setDeckCards(cardsRes.response.cards || []);
      } else {
        setDeckCards([]);
      }
    } catch (err) {
      console.error('Failed to fetch deck details:', err);
    } finally {
      setLoadingDeck(false);
    }
  };

  const fetchAvailableCards = async () => {
    setLoadingAvailable(true);
    try {
      const cardsRes = await cardClient.listCards({ 
        pageSize, 
        pageNumber, 
        ids: [] 
      }, createOptions(token));
      setAvailableCards(cardsRes.response.cards || []);
      setTotalCount(cardsRes.response.totalCount);
    } catch (err) {
      console.error('Failed to fetch available cards:', err);
    } finally {
      setLoadingAvailable(false);
    }
  };

  useEffect(() => {
    fetchDeckAndCards();
  }, [deckId, token]);

  useEffect(() => {
    fetchAvailableCards();
  }, [token, pageNumber]);

  const handleAddCard = async (cardId: string) => {
    try {
      await deckClient.addCardToDeck({ deckId, cardId }, createOptions(token));
      fetchDeckAndCards();
    } catch (err: any) {
      alert(`Failed to add card: ${err.message || err}`);
    }
  };

  const handleRemoveCard = async (cardId: string) => {
    try {
      await deckClient.removeCardFromDeck({ deckId, cardId }, createOptions(token));
      fetchDeckAndCards();
    } catch (err: any) {
      alert(`Failed to remove card: ${err.message || err}`);
    }
  };

  if (!deck) {
    return (
      <div className="max-w-6xl mx-auto p-4 py-12 flex justify-center items-center min-h-[500px]">
        {loadingDeck ? (
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
        ) : (
          <div className="text-center">
            <h2 className="text-3xl font-black text-white mb-4">Deck Not Found</h2>
            <button onClick={onBack} className="text-primary hover:text-white transition-colors flex items-center justify-center gap-2">
              <ArrowLeft size={20} /> GO BACK
            </button>
          </div>
        )}
      </div>
    );
  }

  const isContributor = user && (deck.ownerId === user.id || (deck.contributors || []).includes(user.id));
  const totalPages = Math.ceil(totalCount / pageSize);

  return (
    <div className="max-w-6xl mx-auto p-4 py-12">
      <header className="flex justify-between items-start mb-12">
        <div>
          <button 
            onClick={onBack}
            className="mb-4 text-gray-400 hover:text-white transition-colors flex items-center gap-2 font-bold uppercase tracking-widest text-xs"
          >
            <ArrowLeft size={16} /> Back to Decks
          </button>
          <h1 className="text-5xl font-black tracking-tighter text-white">{deck.name}</h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm mt-2">
            Managing cards for this deck
          </p>
          {!isContributor && (
            <div className="mt-4 p-4 bg-red-500/10 border border-red-500/20 rounded-xl">
              <p className="text-red-400 text-xs font-black uppercase tracking-widest">
                Read Only Access: You are not the owner or a contributor.
              </p>
            </div>
          )}
        </div>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Left Column: Cards in Deck */}
        <div className="glass p-8 rounded-[2.5rem] border border-white/5 shadow-2xl relative min-h-[500px]">
          {loadingDeck && (
             <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center rounded-[2.5rem]">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
             </div>
          )}
          
          <div className="mb-8">
            <h2 className="text-2xl font-black text-white tracking-tight">Cards In Deck ({(deck.cardIds || []).length})</h2>
            <p className="text-gray-500 text-xs font-bold uppercase tracking-widest">Included in {deck.name}</p>
          </div>

          <div className="space-y-4">
            {deckCards.length === 0 ? (
              <div className="py-12 text-center border-2 border-dashed border-white/10 rounded-2xl">
                <p className="text-gray-500 font-bold italic">This deck is empty.</p>
              </div>
            ) : (
              deckCards.map(card => (
                <div 
                  key={card.id} 
                  onClick={() => isContributor && handleRemoveCard(card.id)}
                  className={`p-4 rounded-xl border transition-all ${
                    card.color === 1 
                      ? 'bg-black text-white border-white/10 shadow-md hover:border-red-500/50' 
                      : 'bg-white text-black border-black/5 shadow-md hover:border-red-500/50'
                  } flex justify-between items-center group ${isContributor ? 'cursor-pointer' : ''}`}
                >
                  <p className="font-bold leading-tight pr-4 text-sm">{card.text}</p>
                  {isContributor && (
                    <div 
                      className="text-gray-400 group-hover:text-red-500 transition-colors opacity-0 group-hover:opacity-100 flex-shrink-0"
                      title="Remove from deck"
                    >
                      <Trash2 size={16} />
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        </div>

        {/* Right Column: Available Cards */}
        <div className="glass p-8 rounded-[2.5rem] border border-white/5 shadow-2xl relative min-h-[500px]">
          {loadingAvailable && (
             <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center rounded-[2.5rem]">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
             </div>
          )}

          <div className="mb-8">
            <h2 className="text-2xl font-black text-white tracking-tight">Available Cards</h2>
            <p className="text-gray-500 text-xs font-bold uppercase tracking-widest">Add to your deck</p>
          </div>

          <div className="space-y-4 mb-8">
            {availableCards.map(card => {
              const isInDeck = (deck.cardIds || []).includes(card.id);
              const canAdd = isContributor && !isInDeck;
              return (
                <div 
                  key={card.id} 
                  onClick={() => canAdd && handleAddCard(card.id)}
                  className={`p-4 rounded-xl border transition-all group ${
                    isInDeck ? 'opacity-50 grayscale border-white/5' : 
                    card.color === 1 
                      ? 'bg-black text-white border-white/10 shadow-md hover:border-primary/50' 
                      : 'bg-white text-black border-black/5 shadow-md hover:border-primary/50'
                  } flex justify-between items-center ${canAdd ? 'cursor-pointer' : ''}`}
                >
                  <p className="font-bold leading-tight pr-4 text-sm">{card.text}</p>
                  {isContributor && !isInDeck && (
                    <div 
                      className="bg-primary/10 group-hover:bg-primary text-primary group-hover:text-background p-2 rounded-lg transition-colors flex-shrink-0"
                      title="Add to deck"
                    >
                      <Plus size={16} />
                    </div>
                  )}
                  {isInDeck && (
                    <span className="text-xs font-black tracking-widest uppercase text-gray-500 flex-shrink-0">
                      Added
                    </span>
                  )}
                </div>
              );
            })}
          </div>

          {totalPages > 1 && (
            <div className="flex justify-center items-center gap-4 border-t border-white/10 pt-6">
              <button
                onClick={() => setPageNumber(p => Math.max(1, p - 1))}
                disabled={pageNumber === 1 || loadingAvailable}
                className="px-4 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white text-xs font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
              >
                PREV
              </button>
              <span className="text-gray-400 font-bold text-xs">{pageNumber} / {totalPages}</span>
              <button
                onClick={() => setPageNumber(p => Math.min(totalPages, p + 1))}
                disabled={pageNumber === totalPages || loadingAvailable}
                className="px-4 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white text-xs font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
              >
                NEXT
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
