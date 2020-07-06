var colors = ["blue", "red", "yellow", "green"];
var statuses = {
    0: "NEW",
    1: "ONGOING",
    2: "COMPLETED",
};

class ActiveGameTeam extends React.Component {

    render() {

        const scores = this.props.scores;
        let score;
        if (this.props.scored) {
            score = <h2 className="font-weight-bolder">{scores[this.props.index]}</h2>;
        }
        var color = "bg-" + colors[this.props.index];
        

        let out = "font-weight-bolder";
        var myplace = this.props.places[this.props.index];
        if (myplace < 0 ) {
            myplace = (myplace+1)*(-1)
        }
        if (this.props.status == 2) {
            myplace += 1;
        }
        var placeSpan;
        if (myplace!=0) {
            out += " lost";
            placeSpan = <h2><span class="badge badge-dark">{myplace}</span></h2>
        }
        if (this.props.status == 2 && myplace < Math.ceil(this.props.places.length/2)) {
            out += "winner"
        }

        let image;
        if (this.props.team.image) {
            image = <img src={this.props.team.image}  class="card-img-top"></img>
            color = ""; 
        }

        var colWidth = "col col-" + Math.floor(12/this.props.places.length);



        return (
        <div className={colWidth} key={this.props.team.id}>
            <div className={["card", "team", color].join(' ')}>
                {image}
                <div className="card-body text-center" >
                    <h1 className="font-weight-bolder">{this.props.team.name}</h1>
                    {score}
                    {placeSpan}
                </div>
            </div>
        </div>
        );
    }
}


class ActiveGame extends React.Component {
    render() {
        
        const scores = this.props.data.game.scores;
        const scored = this.props.data.scored;
        const places = this.props.data.game.places;
        const status = this.props.data.game.status;
        const maxSize = this.props.data.game.maxTeams;
        const listItems = this.props.data.game.teams.map(function(team, index) {
            return (<ActiveGameTeam team={team} scores={scores} index={index} places={places} scored={scored} status={status} maxSize={maxSize} />);
        });
        return (
            <div className="container-fluid">
                <div className="row">
                {listItems}
                </div>
            </div>
        );
    }
    
}    



function CreateActiveGame(data, mount) {
    ReactDOM.render(
        <ActiveGame data={data} /> ,
        mount
    );
}

class QueuedGameTeam extends React.Component {
    render() {
        let score;
        if (this.props.scored) {
            score = <span>{this.props.scores[this.props.index]}</span>;
        }
        var myplace = this.props.places[this.props.index];
        if (myplace < 0 ) {
            myplace = (myplace+1)*(-1)
        }
        if (this.props.status == 2) {
            myplace += 1;
        }
        
        let place;
        var placeSpan
        if (myplace!=0) {
            place = "lost";
            placeSpan = <span class="badge badge-dark">{myplace}</span>
        }
        if (this.props.status == 2 && myplace < Math.ceil(this.props.places.length/2)) {
            place = "winner"
        }
        let image;
        if (this.props.team.image) {
            image = <img src={this.props.team.image}  class="img-thumbnail"></img>
        }

        return (
            <li className="list-group-item d-flex justify-content-between align-items-center queued-teams"  key={this.props.team.id}>
                {image}
                <span>{this.props.team.name}</span>
                {score}
                {placeSpan}
            </li>
        );
    };
}

class QueuedGame extends React.Component {
    render() {
        const scores = this.props.game.scores;
        const places = this.props.game.places;
        var gameStatus = this.props.game.status;
        const scored = this.props.scored;
        
        const listItems = this.props.game.teams.map(function(team, index) {
            return (<QueuedGameTeam team={team} index={index} scores={scores} places={places} scored={scored} status={gameStatus}></QueuedGameTeam>);
        });

        let color = "bg-blue";
        if (this.props.game.arena.id) {
            color = "bg-" + colors[this.props.game.arena.id-1];
        }
        let status;
        switch(this.props.game.status) {
            case 1:
                status = <span className="badge badge-dark float-right">ONGOING</span>;
                break;
            case 2:
                status = <span className="badge badge-dark float-right">COMPLETED</span>;
                break;
        }
        
        return (
            <div className="card border-dark mb-3" key={this.props.index}>
                <div className={["card-header text-center", color].join(' ')}>
                <a href={"/arena/"+ this.props.game.arena.name.replace(" ", "").toLowerCase()} class="text-dark font-weight-bolder stretched-link">{this.props.game.arena.name}</a>
                {status}
                </div>
                <div className="card-body text-dark">
                    <ul className="list-group list-group-flush">
                        {listItems}
                    </ul>
                </div>
            </div>
        );
    }
}

class GameQueue extends React.Component {
    
    render() {


        const listItems = this.props.data.map(function(data, index) {
            /*
            if (data.game.status != 2) { // Only show upcoming or current games
                return (<QueuedGame game={data.game} scored={data.scored}  />);
            }
            */
           return (<QueuedGame game={data.game} scored={data.scored} index={index}  />);
        });

        return (
        <div className="queue">
            <div className="queue-inner">
                {listItems}
            </div>
        </div>
        );
    }
}

function CreateGameQueue(data, mount, callback) {
    var queue = ReactDOM.render(
        <GameQueue data={data} callback={callback} /> ,
        mount,
        callback
    );
    return queue
}

class LeaderboardTeam extends React.Component {
    render() {
        var place = this.props.team.place;
        let image;
        if (this.props.team.image) {
            image = <img src={this.props.team.image}  class="img-thumbnail"></img>
        }
        const score = this.props.team.score;
        const name = this.props.team.name;
        const id = this.props.team.id;
        return (
            <a href={"/team/"+ name} class="text-secondary">
                <li className="list-group-item d-flex justify-content-between align-items-center leaderboard-team" key={id}>
                    <span class="font-weight-bolder">{place}.  {image}</span>
                    
                    <span className="leaderboard-player"> {name} </span>
                    <span className="leaderboard-score badge badge-secondary badge-pill">{score}</span>
                </li>
            </a>
        );
    }
}

class Leaderboard extends React.Component {

    render() {

        const listItems = this.props.data.map(function(team, index) {
            return (<LeaderboardTeam team={team} index={index}  />);
        });
        return (
            <ol class="list-group list-group-flush">
                {listItems}
            </ol>
        );
    }
}

function CreateLeaderboard(data, mount) {
    ReactDOM.render(
        <Leaderboard data={data} /> ,
        mount
    );
}