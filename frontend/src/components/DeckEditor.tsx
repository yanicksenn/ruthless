import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Plus, Trash2, ArrowLeft, Search, UserPlus, Users } from 'lucide-react';
import { cardClient, deckClient, createOptions } from '../api/client';
import { Card, Deck, CardColor } from '../api/ruthless';
import { InviteFriendDialog } from './InviteFriendDialog';

interface DeckEditorProps {
  deckId: string;
  initialTab?: string;
  onBack: () => void;
}

export const DeckEditor: React.FC<DeckEditorProps> = ({ deckId, onBack, initialTab = 'cards' }) => {
  const { token, user } = useAuth();
  const [deck, setDeck] = useState<Deck | null>(null);
  const [deckCards, setDeckCards] = useState<Card[]>([]);
  const [availableCards, setAvailableCards] = useState<Card[]>([]);
  
  const [loadingDeck, setLoadingDeck] = useState(true);
  const [loadingAvailable, setLoadingAvailable] = useState(true);
  
  const [deckPage, setDeckPage] = useState(1);
  const [deckTotal, setDeckTotal] = useState(0);
  const [deckFilter, setDeckFilter] = useState('');

  const [availPage, setAvailPage] = useState(1);
  const [availTotal, setAvailTotal] = useState(0);
  const [availFilter, setAvailFilter] = useState('');
  const [excludeInDeck, setExcludeInDeck] = useState(true);

  const pageSize = 12;

  const [activeTab, setActiveTab] = useState(initialTab);
  const [isAddContributorModalOpen, setIsAddContributorModalOpen] = useState(false);

  // Sync tab with URL
  useEffect(() => {
    const newPath = `/library/decks/${deckId}/${activeTab}`;
    if (window.location.pathname !== newPath) {
      window.history.pushState(null, '', newPath);
    }
  }, [activeTab, deckId]);

  const fetchDeckDetails = async () => {
    try {
      const deckRes = await deckClient.getDeck({ id: deckId }, createOptions(token));
      setDeck(deckRes.response);
    } catch (err) {
      console.error('Failed to fetch deck details:', err);
    }
  };

  const fetchDeckCards = async () => {
    setLoadingDeck(true);
    try {
      const cardsRes = await cardClient.listCards({ 
        pageSize, 
        pageNumber: deckPage, 
        ids: [],
        filter: deckFilter,
        includeDeckIds: [deckId],
        color: CardColor.UNSPECIFIED,
        excludeDeckIds: [],
      }, createOptions(token));
      setDeckCards(cardsRes.response.cards || []);
      setDeckTotal(cardsRes.response.totalCount);
    } catch (err) {
      console.error('Failed to fetch deck cards:', err);
    } finally {
      setLoadingDeck(false);
    }
  };

  const fetchAvailableCards = async () => {
    setLoadingAvailable(true);
    try {
      const cardsRes = await cardClient.listCards({ 
        pageSize, 
        pageNumber: availPage, 
        ids: [],
        filter: availFilter,
        includeDeckIds: [],
        color: CardColor.UNSPECIFIED,
        excludeDeckIds: excludeInDeck ? [deckId] : [],
      }, createOptions(token));
      setAvailableCards(cardsRes.response.cards || []);
      setAvailTotal(cardsRes.response.totalCount);
    } catch (err) {
      console.error('Failed to fetch available cards:', err);
    } finally {
      setLoadingAvailable(false);
    }
  };

  useEffect(() => {
    fetchDeckDetails();
  }, [deckId, token]);

  useEffect(() => {
    fetchDeckCards();
  }, [deckId, token, deckPage, deckFilter]);

  // Permissions
  const isContributor = !!(user && deck && (deck.ownerId === user.id || (deck.contributors || []).includes(user.id)));
  const isOwner = !!(user && deck && deck.ownerId === user.id);

  const getContributionCount = (userId: string) => {
    if (!deck || !deck.cardContributorIds) return 0;
    return Object.values(deck.cardContributorIds).filter(id => id === userId).length;
  };

  useEffect(() => {
    if (isContributor) {
      fetchAvailableCards();
    }
  }, [token, availPage, availFilter, excludeInDeck, isContributor]);

  // Reset to first page when filters change
  useEffect(() => {
    setDeckPage(1);
  }, [deckFilter]);

  useEffect(() => {
    setAvailPage(1);
  }, [availFilter, excludeInDeck]);

  const handleAddCard = async (cardId: string) => {
    try {
      await deckClient.addCardToDeck({ deckId, cardId }, createOptions(token));
      fetchDeckDetails();
      fetchDeckCards();
    } catch (err: any) {
      alert(`Failed to add card: ${err.message || err}`);
    }
  };

  const handleRemoveCard = async (cardId: string) => {
    try {
      await deckClient.removeCardFromDeck({ deckId, cardId }, createOptions(token));
      fetchDeckDetails();
      fetchDeckCards();
    } catch (err: any) {
      alert(`Failed to remove card: ${err.message || err}`);
    }
  };

  const handleAddContributor = async (identifier: string) => {
    try {
      await deckClient.addContributor({ deckId, identifier }, createOptions(token));
      setIsAddContributorModalOpen(false);
      fetchDeckDetails();
    } catch (err: any) {
      alert(`Failed to add contributor: ${err.message || err}`);
    }
  };

  const handleRemoveContributor = async (identifier: string) => {
    try {
      await deckClient.removeContributor({ deckId, identifier }, createOptions(token));
      fetchDeckDetails();
    } catch (err: any) {
      alert(`Failed to remove contributor: ${err.message || err}`);
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

  const deckPages = Math.ceil(deckTotal / pageSize);
  const availPages = Math.ceil(availTotal / pageSize);

  return (
    <div className="max-w-6xl mx-auto p-4 py-12">
      <header className="mb-8">
          <button 
            onClick={onBack}
            className="mb-4 text-gray-400 hover:text-white transition-colors flex items-center gap-2 font-bold uppercase tracking-widest text-xs"
          >
            <ArrowLeft size={16} /> Back to Library
          </button>
          <div className="flex flex-col gap-2">
              <h1 className="text-5xl font-black tracking-tighter text-white">{deck.name}</h1>
              <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">
                {activeTab === 'cards' ? 'Managing cards for this deck' : 'Managing contributors for this deck'}
              </p>
              {!isContributor && (
                <div className="mt-4 p-4 bg-red-500/10 border border-red-500/20 rounded-xl max-w-md">
                  <p className="text-red-400 text-xs font-black uppercase tracking-widest">
                    Read Only Access: You are not the owner or a contributor.
                  </p>
                </div>
              )}
          </div>
      </header>

      {/* Sub-menu */}
      <div className="flex gap-4 mb-8 border-b border-white/10 pb-4">
        <button
          onClick={() => setActiveTab('cards')}
          className={`px-6 py-2 rounded-xl font-black text-sm transition-all tracking-widest uppercase ${
            activeTab === 'cards' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white hover:bg-white/5'
          }`}
        >
          Cards
        </button>
        <button
          onClick={() => setActiveTab('contributors')}
          className={`px-6 py-2 rounded-xl font-black text-sm transition-all tracking-widest uppercase ${
            activeTab === 'contributors' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white hover:bg-white/5'
          }`}
        >
          Contributors
        </button>
      </div>

      {activeTab === 'cards' ? (
        <div className={`grid grid-cols-1 ${isContributor ? 'lg:grid-cols-2' : 'max-w-3xl mx-auto'} gap-8`}>
          {/* Left Column: Cards in Deck */}
          <div className="glass p-8 rounded-[2.5rem] border border-white/5 shadow-2xl relative min-h-[500px]">
            {loadingDeck && (
               <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center rounded-[2.5rem]">
                  <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
               </div>
            )}
            
            <div className="mb-8 space-y-4">
              <div>
                <h2 className="text-2xl font-black text-white tracking-tight">Cards In Deck ({deckTotal})</h2>
                <p className="text-gray-500 text-xs font-bold uppercase tracking-widest">Included in {deck.name}</p>
              </div>

              <div className="relative group/search text-white">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 group-hover/search:text-primary transition-colors" size={16} />
                <input
                  type="text"
                  placeholder="Filter deck cards..."
                  value={deckFilter}
                  onChange={(e) => setDeckFilter(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl py-2 pl-12 pr-4 text-sm font-medium placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/30 transition-all"
                />
              </div>
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
                        className="h-9 w-9 flex items-center justify-center text-gray-400 group-hover:text-red-500 transition-colors opacity-0 group-hover:opacity-100 flex-shrink-0"
                        title="Remove from deck"
                      >
                        <Trash2 size={16} />
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>

            {deckPages > 1 && (
              <div className="flex justify-center items-center gap-4 border-t border-white/10 pt-6 mt-8 text-white">
                <button
                  onClick={() => setDeckPage(p => Math.max(1, p - 1))}
                  disabled={deckPage === 1 || loadingDeck}
                  className="px-4 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white text-xs font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                >
                  PREV
                </button>
                <span className="text-gray-400 font-bold text-xs">{deckPage} / {deckPages}</span>
                <button
                  onClick={() => setDeckPage(p => Math.min(deckPages, p + 1))}
                  disabled={deckPage === deckPages || loadingDeck}
                  className="px-4 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white text-xs font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                >
                  NEXT
                </button>
              </div>
            )}
          </div>

          {/* Right Column: Available Cards */}
          {isContributor && (
            <div className="glass p-8 rounded-[2.5rem] border border-white/5 shadow-2xl relative min-h-[500px]">
            {loadingAvailable && (
               <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center rounded-[2.5rem]">
                  <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
               </div>
            )}

            <div className="mb-8 space-y-4">
              <div>
                <h2 className="text-2xl font-black text-white tracking-tight">Available Cards</h2>
                <p className="text-gray-500 text-xs font-bold uppercase tracking-widest">Add to your deck</p>
              </div>

              <div className="flex items-center justify-between gap-4">
                <div className="relative group/search text-white flex-1">
                  <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 group-hover/search:text-primary transition-colors" size={16} />
                  <input
                    type="text"
                    placeholder="Filter available cards..."
                    value={availFilter}
                    onChange={(e) => setAvailFilter(e.target.value)}
                    className="w-full bg-white/5 border border-white/10 rounded-xl py-2 pl-12 pr-4 text-sm font-medium placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/30 transition-all"
                  />
                </div>
                
                <button
                  onClick={() => setExcludeInDeck(!excludeInDeck)}
                  className={`flex items-center gap-2 px-4 py-2 rounded-xl border transition-all whitespace-nowrap text-xs font-black uppercase tracking-widest ${
                    excludeInDeck 
                      ? 'bg-primary/20 border-primary/30 text-primary' 
                      : 'bg-white/5 border-white/10 text-gray-500 hover:text-white'
                  }`}
                >
                  <div className={`w-3 h-3 rounded-full border-2 transition-all ${excludeInDeck ? 'bg-primary border-primary' : 'border-gray-600'}`} />
                  Exclude in deck
                </button>
              </div>
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
                        className="h-9 w-9 flex items-center justify-center bg-primary/10 group-hover:bg-primary text-primary group-hover:text-background rounded-lg transition-colors flex-shrink-0"
                        title="Add to deck"
                      >
                        <Plus size={16} />
                      </div>
                    )}
                    {isInDeck && (
                      <div className="h-9 flex items-center justify-center flex-shrink-0">
                        <span className="text-xs font-black tracking-widest uppercase text-gray-500">
                          Added
                        </span>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>

            {availPages > 1 && (
              <div className="flex justify-center items-center gap-4 border-t border-white/10 pt-6">
                <button
                  onClick={() => setAvailPage(p => Math.max(1, p - 1))}
                  disabled={availPage === 1 || loadingAvailable}
                  className="px-4 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white text-xs font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                >
                  PREV
                </button>
                <span className="text-gray-400 font-bold text-xs">{availPage} / {availPages}</span>
                <button
                  onClick={() => setAvailPage(p => Math.min(availPages, p + 1))}
                  disabled={availPage === availPages || loadingAvailable}
                  className="px-4 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white text-xs font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                >
                  NEXT
                </button>
              </div>
            )}
          </div>
          )}
        </div>
      ) : (
        /* Contributors Tab */
        <div className="max-w-2xl mx-auto glass p-8 rounded-[2.5rem] border border-white/5 shadow-2xl">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-bold text-white uppercase tracking-widest flex items-center gap-2">
              <Users size={18} className="text-primary" /> Card Contributors
            </h2>
            {isOwner && (
              <button
                onClick={() => setIsAddContributorModalOpen(true)}
                className="bg-primary hover:bg-primary/80 text-background font-black px-4 py-2 rounded-xl flex items-center gap-2 transition-all transform hover:scale-105 shadow-sm shadow-primary/20 uppercase tracking-widest text-xs"
              >
                <UserPlus size={16} /> Add Contributor
              </button>
            )}
          </div>

          <div className="space-y-6">
            {/* Owner Section */}
            <div>
              <h3 className="text-xs font-black text-primary uppercase tracking-widest mb-3">Owner</h3>
              {deck?.ownerPlayer && (
                 <div className="flex justify-between items-center p-4 bg-primary/5 rounded-2xl border border-primary/20">
                   <span className="font-bold text-white tracking-tight">
                      {deck.ownerPlayer.name}
                      {deck.ownerPlayer.identifier && <span className="text-primary/60 italic ml-1">#{deck.ownerPlayer.identifier}</span>}
                   </span>
                   <span className="text-[10px] font-black bg-primary text-background px-2 py-0.5 rounded-md tracking-tighter uppercase">
                     {getContributionCount(deck.ownerPlayer.id)} Cards • Original Creator
                   </span>
                 </div>
              )}
            </div>

            {/* Contributors Section */}
            <div>
              <h3 className="text-xs font-black text-gray-500 uppercase tracking-widest mb-3">Contributors</h3>
              <div className="space-y-3">
                {deck.contributorPlayers && deck.contributorPlayers.length > 0 ? (
                  deck.contributorPlayers.map(p => (
                    <div key={p.id} className="flex justify-between items-center p-4 bg-white/5 rounded-2xl border border-white/5 group">
                      <div className="flex flex-col">
                        <span className="font-bold text-gray-300 tracking-tight">
                           {p.name}
                           {p.identifier && <span className="text-gray-600 italic ml-1">#{p.identifier}</span>}
                        </span>
                        <span className="text-[10px] font-black text-gray-500 uppercase tracking-widest mt-1">
                          {getContributionCount(p.id)} Cards
                        </span>
                      </div>
                      {isOwner && (
                        <button 
                          onClick={() => p.identifier && handleRemoveContributor(p.identifier)} 
                          className="text-gray-500 hover:text-red-500 p-2 rounded-xl opacity-0 group-hover:opacity-100 transition-all hover:bg-white/10"
                          title="Remove contributor"
                        >
                          <Trash2 size={18} />
                        </button>
                      )}
                    </div>
                  ))
                ) : (
                  <div className="text-center py-8 border-2 border-dashed border-white/5 rounded-2xl">
                    <p className="text-gray-600 text-sm font-bold uppercase tracking-widest italic">No contributors added</p>
                  </div>
                )}
              </div>
            </div>

            {isOwner && (
              <p className="mt-2 text-[10px] text-gray-600 font-bold uppercase tracking-widest">
                Invite others using their unique name and identifier.
              </p>
            )}
          </div>
        </div>
      )}

      <InviteFriendDialog
        isOpen={isAddContributorModalOpen}
        onClose={() => setIsAddContributorModalOpen(false)}
        onAction={handleAddContributor}
        excludeFromDeckId={deckId}
        title="Add Contributor"
        buttonText="Add"
      />
    </div>
  );
};
