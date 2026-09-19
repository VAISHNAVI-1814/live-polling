import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api';
import ShareModal from '../components/ShareModal';
import { 
  PlusCircle, 
  BarChart3, 
  Share2, 
  Trash2, 
  ToggleLeft, 
  ToggleRight, 
  Clock, 
  Vote, 
  CheckCircle2, 
  AlertCircle,
  ExternalLink
} from 'lucide-react';

export default function Dashboard() {
  const [polls, setPolls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedSharePoll, setSelectedSharePoll] = useState(null);

  const fetchPolls = async () => {
    try {
      setLoading(true);
      const data = await api.getMyPolls();
      setPolls(data);
    } catch (err) {
      setError(err.message || 'Failed to load your polls');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPolls();
  }, []);

  const handleToggleStatus = async (poll) => {
    const newStatus = poll.status === 'active' ? 'closed' : 'active';
    try {
      await api.updatePollStatus(poll.id, newStatus);
      setPolls(polls.map(p => p.id === poll.id ? { ...p, status: newStatus } : p));
    } catch (err) {
      alert(err.message || 'Failed to update poll status');
    }
  };

  const handleDelete = async (pollId) => {
    if (!window.confirm('Are you sure you want to delete this poll? All live votes and data will be permanently removed.')) {
      return;
    }
    try {
      await api.deletePoll(pollId);
      setPolls(polls.filter(p => p.id !== pollId));
    } catch (err) {
      alert(err.message || 'Failed to delete poll');
    }
  };

  const totalVotesAcrossPolls = polls.reduce((acc, p) => acc + (p.total_votes || 0), 0);
  const activePollsCount = polls.filter(p => p.status === 'active').length;

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      {/* Top Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">Poll Dashboard</h1>
          <p className="text-sm text-slate-400 mt-1">Manage your active polls and view live audience voting metrics.</p>
        </div>
        <Link
          to="/create-poll"
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-sm rounded-xl shadow-lg shadow-indigo-600/30 transition-all hover:-translate-y-0.5"
        >
          <PlusCircle className="w-4 h-4" />
          Create New Poll
        </Link>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-5 backdrop-blur-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">Total Polls</span>
            <div className="p-2 rounded-xl bg-indigo-500/10 text-indigo-400">
              <Vote className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-white mt-2">{polls.length}</p>
        </div>

        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-5 backdrop-blur-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">Active Polls</span>
            <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-400">
              <CheckCircle2 className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-emerald-400 mt-2">{activePollsCount}</p>
        </div>

        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-5 backdrop-blur-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">Total Live Votes</span>
            <div className="p-2 rounded-xl bg-purple-500/10 text-purple-400">
              <BarChart3 className="w-4 h-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-purple-400 mt-2">{totalVotesAcrossPolls}</p>
        </div>
      </div>

      {/* Main List */}
      {error && (
        <div className="mb-6 p-4 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm flex items-center gap-3">
          <AlertCircle className="w-5 h-5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {loading ? (
        <div className="flex justify-center items-center py-20">
          <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-500"></div>
        </div>
      ) : polls.length === 0 ? (
        <div className="text-center py-16 px-4 bg-slate-900/40 border border-dashed border-slate-800 rounded-2xl">
          <div className="w-14 h-14 rounded-2xl bg-indigo-500/10 text-indigo-400 flex items-center justify-center mx-auto mb-4">
            <Vote className="w-7 h-7" />
          </div>
          <h3 className="text-lg font-semibold text-white">No polls created yet</h3>
          <p className="text-sm text-slate-400 max-w-sm mx-auto mt-1.5 mb-6">
            Get started by launching your first live poll and sharing the interactive link with your audience.
          </p>
          <Link
            to="/create-poll"
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-sm rounded-xl shadow-lg shadow-indigo-600/30 transition-all hover:-translate-y-0.5"
          >
            <PlusCircle className="w-4 h-4" />
            Create Poll Now
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {polls.map((poll) => {
            const isActive = poll.status === 'active';
            const isExpired = poll.expires_at && new Date(poll.expires_at) < new Date();

            return (
              <div
                key={poll.id}
                className="bg-slate-900/90 border border-slate-800/80 hover:border-slate-700/80 rounded-2xl p-5 sm:p-6 transition-all duration-200 backdrop-blur-sm"
              >
                <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                  {/* Poll Info */}
                  <div className="space-y-2 flex-1 min-w-0">
                    <div className="flex flex-wrap items-center gap-2.5">
                      <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold ${
                        isActive && !isExpired
                          ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                          : 'bg-slate-800 text-slate-400 border border-slate-700'
                      }`}>
                        <span className={`w-1.5 h-1.5 rounded-full ${isActive && !isExpired ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`} />
                        {isExpired ? 'Expired' : isActive ? 'Active' : 'Closed'}
                      </span>

                      <span className="font-mono text-xs px-2.5 py-0.5 rounded-full bg-slate-950 text-indigo-400 border border-slate-800">
                        Code: {poll.share_code}
                      </span>

                      {poll.expires_at && (
                        <span className="flex items-center gap-1 text-xs text-slate-400">
                          <Clock className="w-3.5 h-3.5 text-slate-500" />
                          {isExpired ? 'Expired' : `Closes ${new Date(poll.expires_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`}
                        </span>
                      )}
                    </div>

                    <h3 className="text-lg font-semibold text-white truncate">
                      {poll.question}
                    </h3>

                    <div className="flex items-center gap-4 text-xs text-slate-400">
                      <span>{poll.options?.length || 0} Options</span>
                      <span>•</span>
                      <span className="font-mono font-medium text-slate-300">
                        {poll.total_votes || 0} Total Votes
                      </span>
                      <span>•</span>
                      <span>Created {new Date(poll.created_at).toLocaleDateString()}</span>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex flex-wrap items-center gap-2 pt-3 lg:pt-0 border-t lg:border-t-0 border-slate-800">
                    <button
                      onClick={() => setSelectedSharePoll(poll)}
                      className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-xl transition-colors"
                      title="Share link & code"
                    >
                      <Share2 className="w-3.5 h-3.5 text-indigo-400" />
                      Share
                    </button>

                    <Link
                      to={`/poll/${poll.id}/results`}
                      className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 border border-indigo-500/30 rounded-xl transition-colors"
                      title="View live updating results"
                    >
                      <BarChart3 className="w-3.5 h-3.5" />
                      Live Results
                    </Link>

                    <Link
                      to={`/poll/${poll.share_code}`}
                      target="_blank"
                      className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors"
                      title="Test vote in new tab"
                    >
                      <ExternalLink className="w-4 h-4" />
                    </Link>

                    <button
                      onClick={() => handleToggleStatus(poll)}
                      className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors"
                      title={isActive ? 'Close Poll' : 'Reopen Poll'}
                    >
                      {isActive ? (
                        <ToggleRight className="w-5 h-5 text-emerald-400" />
                      ) : (
                        <ToggleLeft className="w-5 h-5 text-slate-500" />
                      )}
                    </button>

                    <button
                      onClick={() => handleDelete(poll.id)}
                      className="p-2 text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 rounded-xl transition-colors"
                      title="Delete poll"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Share Modal */}
      <ShareModal
        poll={selectedSharePoll}
        isOpen={!!selectedSharePoll}
        onClose={() => setSelectedSharePoll(null)}
      />
    </div>
  );
}
