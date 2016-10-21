%%%-------------------------------------------------------------------
%%% This file is part of the CernVM File System.
%%%
%%% @doc cvmfs_lease top level supervisor.
%%%
%%% @end
%%%-------------------------------------------------------------------

-module(cvmfs_auth_sup).

-behaviour(supervisor).

%% API
-export([start_link/1]).

%% Supervisor callbacks
-export([init/1]).

-define(SERVER, ?MODULE).

%%====================================================================
%% API functions
%%====================================================================

start_link(Args) ->
    supervisor:start_link({local, ?SERVER}, ?MODULE, Args).

%%====================================================================
%% Supervisor callbacks
%%====================================================================

%% Child :: {Id,StartFunc,Restart,Shutdown,Type,Modules}
init(Args) ->
    SupervisorSpecs = #{strategy => one_for_all,
                        intensity => 5,
                        period => 5},
    CvmfsAuthMainSpecs = #{id => cvmfs_auth,
                             start => {cvmfs_auth, start_link, [Args]},
                             restart => permanent,
                             shutdown => 2000,
                             type => worker,
                             modules => [cvmfs_auth]},
    {ok, {SupervisorSpecs, [CvmfsAuthMainSpecs]}}.

%%====================================================================
%% Internal functions
%%====================================================================