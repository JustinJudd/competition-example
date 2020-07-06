"use strict";

function _instanceof(left, right) { if (right != null && typeof Symbol !== "undefined" && right[Symbol.hasInstance]) { return right[Symbol.hasInstance](left); } else { return left instanceof right; } }

function _typeof(obj) { if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

function _classCallCheck(instance, Constructor) { if (!_instanceof(instance, Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }

function _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }

function _possibleConstructorReturn(self, call) { if (call && (_typeof(call) === "object" || typeof call === "function")) { return call; } return _assertThisInitialized(self); }

function _assertThisInitialized(self) { if (self === void 0) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return self; }

function _getPrototypeOf(o) { _getPrototypeOf = Object.setPrototypeOf ? Object.getPrototypeOf : function _getPrototypeOf(o) { return o.__proto__ || Object.getPrototypeOf(o); }; return _getPrototypeOf(o); }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function"); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, writable: true, configurable: true } }); if (superClass) _setPrototypeOf(subClass, superClass); }

function _setPrototypeOf(o, p) { _setPrototypeOf = Object.setPrototypeOf || function _setPrototypeOf(o, p) { o.__proto__ = p; return o; }; return _setPrototypeOf(o, p); }

var colors = ["blue", "red", "yellow", "green"];
var statuses = {
  0: "NEW",
  1: "ONGOING",
  2: "COMPLETED"
};

var ActiveGameTeam =
/*#__PURE__*/
function (_React$Component) {
  _inherits(ActiveGameTeam, _React$Component);

  function ActiveGameTeam() {
    _classCallCheck(this, ActiveGameTeam);

    return _possibleConstructorReturn(this, _getPrototypeOf(ActiveGameTeam).apply(this, arguments));
  }

  _createClass(ActiveGameTeam, [{
    key: "render",
    value: function render() {
      var scores = this.props.scores;
      var score;

      if (this.props.scored) {
        score = React.createElement("h2", {
          className: "font-weight-bolder"
        }, scores[this.props.index]);
      }

      var color = "bg-" + colors[this.props.index];
      var out = "font-weight-bolder";
      var myplace = this.props.places[this.props.index];

      if (myplace < 0) {
        myplace = (myplace + 1) * -1;
      }

      if (this.props.status == 2) {
        myplace += 1;
      }

      var placeSpan;

      if (myplace != 0) {
        out += " lost";
        placeSpan = React.createElement("h2", null, React.createElement("span", {
          class: "badge badge-dark"
        }, myplace));
      }

      if (this.props.status == 2 && myplace < Math.ceil(this.props.places.length / 2)) {
        out += "winner";
      }

      var image;

      if (this.props.team.image) {
        image = React.createElement("img", {
          src: this.props.team.image,
          class: "card-img-top"
        });
        color = "";
      }

      var colWidth = "col col-" + Math.floor(12 / this.props.places.length);
      return React.createElement("div", {
        className: colWidth,
        key: this.props.team.id
      }, React.createElement("div", {
        className: ["card", "team", color].join(' ')
      }, image, React.createElement("div", {
        className: "card-body text-center"
      }, React.createElement("h1", {
        className: "font-weight-bolder"
      }, this.props.team.name), score, placeSpan)));
    }
  }]);

  return ActiveGameTeam;
}(React.Component);

var ActiveGame =
/*#__PURE__*/
function (_React$Component2) {
  _inherits(ActiveGame, _React$Component2);

  function ActiveGame() {
    _classCallCheck(this, ActiveGame);

    return _possibleConstructorReturn(this, _getPrototypeOf(ActiveGame).apply(this, arguments));
  }

  _createClass(ActiveGame, [{
    key: "render",
    value: function render() {
      var scores = this.props.data.game.scores;
      var scored = this.props.data.scored;
      var places = this.props.data.game.places;
      var status = this.props.data.game.status;
      var maxSize = this.props.data.game.maxTeams;
      var listItems = this.props.data.game.teams.map(function (team, index) {
        return React.createElement(ActiveGameTeam, {
          team: team,
          scores: scores,
          index: index,
          places: places,
          scored: scored,
          status: status,
          maxSize: maxSize
        });
      });
      return React.createElement("div", {
        className: "container-fluid"
      }, React.createElement("div", {
        className: "row"
      }, listItems));
    }
  }]);

  return ActiveGame;
}(React.Component);

function CreateActiveGame(data, mount) {
  ReactDOM.render(React.createElement(ActiveGame, {
    data: data
  }), mount);
}

var QueuedGameTeam =
/*#__PURE__*/
function (_React$Component3) {
  _inherits(QueuedGameTeam, _React$Component3);

  function QueuedGameTeam() {
    _classCallCheck(this, QueuedGameTeam);

    return _possibleConstructorReturn(this, _getPrototypeOf(QueuedGameTeam).apply(this, arguments));
  }

  _createClass(QueuedGameTeam, [{
    key: "render",
    value: function render() {
      var score;

      if (this.props.scored) {
        score = React.createElement("span", null, this.props.scores[this.props.index]);
      }

      var myplace = this.props.places[this.props.index];

      if (myplace < 0) {
        myplace = (myplace + 1) * -1;
      }

      if (this.props.status == 2) {
        myplace += 1;
      }

      var place;
      var placeSpan;

      if (myplace != 0) {
        place = "lost";
        placeSpan = React.createElement("span", {
          class: "badge badge-dark"
        }, myplace);
      }

      if (this.props.status == 2 && myplace < Math.ceil(this.props.places.length / 2)) {
        place = "winner";
      }

      var image;

      if (this.props.team.image) {
        image = React.createElement("img", {
          src: this.props.team.image,
          class: "img-thumbnail"
        });
      }

      return React.createElement("li", {
        className: "list-group-item d-flex justify-content-between align-items-center queued-teams",
        key: this.props.team.id
      }, image, React.createElement("span", null, this.props.team.name), score, placeSpan);
    }
  }]);

  return QueuedGameTeam;
}(React.Component);

var QueuedGame =
/*#__PURE__*/
function (_React$Component4) {
  _inherits(QueuedGame, _React$Component4);

  function QueuedGame() {
    _classCallCheck(this, QueuedGame);

    return _possibleConstructorReturn(this, _getPrototypeOf(QueuedGame).apply(this, arguments));
  }

  _createClass(QueuedGame, [{
    key: "render",
    value: function render() {
      var scores = this.props.game.scores;
      var places = this.props.game.places;
      var gameStatus = this.props.game.status;
      var scored = this.props.scored;
      var listItems = this.props.game.teams.map(function (team, index) {
        return React.createElement(QueuedGameTeam, {
          team: team,
          index: index,
          scores: scores,
          places: places,
          scored: scored,
          status: gameStatus
        });
      });
      var color = "bg-blue";

      if (this.props.game.arena.id) {
        color = "bg-" + colors[this.props.game.arena.id - 1];
      }

      var status;

      switch (this.props.game.status) {
        case 1:
          status = React.createElement("span", {
            className: "badge badge-dark float-right"
          }, "ONGOING");
          break;

        case 2:
          status = React.createElement("span", {
            className: "badge badge-dark float-right"
          }, "COMPLETED");
          break;
      }

      return React.createElement("div", {
        className: "card border-dark mb-3",
        key: this.props.index
      }, React.createElement("div", {
        className: ["card-header text-center", color].join(' ')
      }, React.createElement("a", {
        href: "/arena/" + this.props.game.arena.name.replace(" ", "").toLowerCase(),
        class: "text-dark font-weight-bolder stretched-link"
      }, this.props.game.arena.name), status), React.createElement("div", {
        className: "card-body text-dark"
      }, React.createElement("ul", {
        className: "list-group list-group-flush"
      }, listItems)));
    }
  }]);

  return QueuedGame;
}(React.Component);

var GameQueue =
/*#__PURE__*/
function (_React$Component5) {
  _inherits(GameQueue, _React$Component5);

  function GameQueue() {
    _classCallCheck(this, GameQueue);

    return _possibleConstructorReturn(this, _getPrototypeOf(GameQueue).apply(this, arguments));
  }

  _createClass(GameQueue, [{
    key: "render",
    value: function render() {
      var listItems = this.props.data.map(function (data, index) {
        /*
        if (data.game.status != 2) { // Only show upcoming or current games
            return (<QueuedGame game={data.game} scored={data.scored}  />);
        }
        */
        return React.createElement(QueuedGame, {
          game: data.game,
          scored: data.scored,
          index: index
        });
      });
      return React.createElement("div", {
        className: "queue"
      }, React.createElement("div", {
        className: "queue-inner"
      }, listItems));
    }
  }]);

  return GameQueue;
}(React.Component);

function CreateGameQueue(data, mount, callback) {
  var queue = ReactDOM.render(React.createElement(GameQueue, {
    data: data,
    callback: callback
  }), mount, callback);
  return queue;
}

var LeaderboardTeam =
/*#__PURE__*/
function (_React$Component6) {
  _inherits(LeaderboardTeam, _React$Component6);

  function LeaderboardTeam() {
    _classCallCheck(this, LeaderboardTeam);

    return _possibleConstructorReturn(this, _getPrototypeOf(LeaderboardTeam).apply(this, arguments));
  }

  _createClass(LeaderboardTeam, [{
    key: "render",
    value: function render() {
      var place = this.props.team.place;
      var image;

      if (this.props.team.image) {
        image = React.createElement("img", {
          src: this.props.team.image,
          class: "img-thumbnail"
        });
      }

      var score = this.props.team.score;
      var name = this.props.team.name;
      var id = this.props.team.id;
      return React.createElement("a", {
        href: "/team/" + name,
        class: "text-secondary"
      }, React.createElement("li", {
        className: "list-group-item d-flex justify-content-between align-items-center leaderboard-team",
        key: id
      }, React.createElement("span", {
        class: "font-weight-bolder"
      }, place, ".  ", image), React.createElement("span", {
        className: "leaderboard-player"
      }, " ", name, " "), React.createElement("span", {
        className: "leaderboard-score badge badge-secondary badge-pill"
      }, score)));
    }
  }]);

  return LeaderboardTeam;
}(React.Component);

var Leaderboard =
/*#__PURE__*/
function (_React$Component7) {
  _inherits(Leaderboard, _React$Component7);

  function Leaderboard() {
    _classCallCheck(this, Leaderboard);

    return _possibleConstructorReturn(this, _getPrototypeOf(Leaderboard).apply(this, arguments));
  }

  _createClass(Leaderboard, [{
    key: "render",
    value: function render() {
      var listItems = this.props.data.map(function (team, index) {
        return React.createElement(LeaderboardTeam, {
          team: team,
          index: index
        });
      });
      return React.createElement("ol", {
        class: "list-group list-group-flush"
      }, listItems);
    }
  }]);

  return Leaderboard;
}(React.Component);

function CreateLeaderboard(data, mount) {
  ReactDOM.render(React.createElement(Leaderboard, {
    data: data
  }), mount);
}