import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { Plus, Trash2, Clock, Sparkles, AlertCircle, CheckCircle2, ArrowRight } from 'lucide-react';

export default function CreatePoll() {
  const [question, setQuestion] = useState('');
  const [options, setOptions] = useState(['', '']);
  const [expiresInMinutes, setExpiresInMinutes] = useState(0); // 0 = never
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const navigate = useNavigate();

  const handleAddOption = () => {
    if (options.length >= 10) return;
    setOptions([...options, '']);
  };

  const handleRemoveOption = (index) => {
    if (options.length <= 2) return;
    setOptions(options.filter((_, i) => i !== index));
  };

  const handleOptionChange = (index, value) => {
    const updated = [...options];
    updated[index] = value;
    setOptions(updated);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const trimmedQuestion = question.trim();
    if (!trimmedQuestion || trimmedQuestion.length < 5) {
      setError('Question must be at least 5 characters long');
      return;
    }

    const trimmedOptions = options.map(o => o.trim()).filter(Boolean);
    if (trimmedOptions.length < 2) {
      setError('Please provide at least 2 non-empty options');
      return;
    }

    const uniqueSet = new Set(trimmedOptions.map(o => o.toLowerCase()));
    if (uniqueSet.size !== trimmedOptions.length) {
      setError('Options must be unique (no duplicates)');
      return;
    }

    try {
      setIsSubmitting(true);
      const poll = await api.createPoll({
        question: trimmedQuestion,
        options: trimmedOptions,
        expires_in_minutes: Number(expiresInMinutes),
      });
      navigate(`/poll/${poll.id}/results`);
    } catch (err) {
      setError(err.message || 'Failed to create poll');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <div className="mb-8">
        <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">Create a Live Poll</h1>
        <p className="text-sm text-slate-400 mt-1">Configure your question and choices. Your live link will be ready immediately.</p>
      </div>

      <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 sm:p-8 shadow-xl backdrop-blur-sm">
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-sm flex items-center gap-3">
            <AlertCircle className="w-5 h-5 flex-shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Question */}
          <div>
            <label className="block text-sm font-medium text-slate-200 mb-2">
              Poll Question <span className="text-rose-400">*</span>
            </label>
            <input
              type="text"
              required
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              placeholder="e.g. What is your favorite programming language?"
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
            />
          </div>

          {/* Options */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="block text-sm font-medium text-slate-200">
                Poll Options <span className="text-rose-400">*</span>
              </label>
              <span className="text-xs text-slate-400">
                {options.length}/10 options
              </span>
            </div>

            <div className="space-y-3">
              {options.map((opt, idx) => (
                <div key={idx} className="flex items-center gap-2">
                  <span className="w-6 text-center text-xs font-mono text-slate-500 font-bold">
                    {idx + 1}.
                  </span>
                  <input
                    type="text"
                    required
                    value={opt}
                    onChange={(e) => handleOptionChange(idx, e.target.value)}
                    placeholder={`Option ${idx + 1}`}
                    className="flex-1 bg-slate-950 border border-slate-800 rounded-xl px-4 py-2.5 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
                  />
                  {options.length > 2 && (
                    <button
                      type="button"
                      onClick={() => handleRemoveOption(idx)}
                      className="p-2.5 text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 rounded-xl transition-colors"
                      title="Remove option"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  )}
                </div>
              ))}
            </div>

            {options.length < 10 && (
              <button
                type="button"
                onClick={handleAddOption}
                className="mt-3 inline-flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold text-indigo-400 hover:text-indigo-300 bg-indigo-500/10 hover:bg-indigo-500/20 border border-indigo-500/20 rounded-xl transition-all"
              >
                <Plus className="w-3.5 h-3.5" />
                Add Another Option
              </button>
            )}
          </div>

          {/* Expiration Preset */}
          <div>
            <label className="block text-sm font-medium text-slate-200 mb-2 flex items-center gap-1.5">
              <Clock className="w-4 h-4 text-slate-400" />
              Poll Duration
            </label>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
              {[
                { label: 'Never Expires', value: 0 },
                { label: '1 Hour', value: 60 },
                { label: '24 Hours', value: 1440 },
                { label: '7 Days', value: 10080 },
              ].map((preset) => (
                <button
                  key={preset.value}
                  type="button"
                  onClick={() => setExpiresInMinutes(preset.value)}
                  className={`py-2.5 px-3 rounded-xl text-xs font-medium border transition-all ${
                    expiresInMinutes === preset.value
                      ? 'bg-indigo-600/20 border-indigo-500 text-indigo-300 shadow-sm shadow-indigo-500/20'
                      : 'bg-slate-950 border-slate-800 text-slate-400 hover:text-white hover:border-slate-700'
                  }`}
                >
                  {preset.label}
                </button>
              ))}
            </div>
          </div>

          {/* Live Preview Card */}
          {question && (
            <div className="mt-8 pt-6 border-t border-slate-800">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider block mb-3 flex items-center gap-1.5">
                <Sparkles className="w-3.5 h-3.5 text-indigo-400" />
                Voter Preview
              </span>
              <div className="bg-slate-950 border border-slate-800 rounded-2xl p-5">
                <h4 className="font-semibold text-white text-base mb-3">{question}</h4>
                <div className="space-y-2">
                  {options.filter(Boolean).map((opt, i) => (
                    <div key={i} className="p-3 rounded-xl bg-slate-900 border border-slate-800 text-xs text-slate-300 flex items-center gap-2">
                      <div className="w-3.5 h-3.5 rounded-full border border-slate-600"></div>
                      <span>{opt}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Submit */}
          <div className="pt-4 border-t border-slate-800 flex justify-end">
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full sm:w-auto px-6 py-3 rounded-xl bg-indigo-600 hover:bg-indigo-500 active:scale-[0.99] text-white text-sm font-semibold shadow-lg shadow-indigo-600/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSubmitting ? (
                <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent" />
              ) : (
                <>
                  Publish Live Poll
                  <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
