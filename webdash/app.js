var angular = require('angular')
var angular_route = require('angular-route')
var ngtable = require('ng-table')
var mdl = require('material-design-lite')
var app = angular.module('TESApp', ['ngRoute', 'ngTable']);

function shortID(longID) {
  return longID.split('-')[0];
}

function listAllTasks($http, tasks, page) {
  if (!tasks) {
    tasks = []
  }

  var url = "/v1/tasks";
  if (page) {
    url += "?page_token=" + page;
  }

  return $http.get(url).then(function(result) {
    Array.prototype.push.apply(tasks, result.data.tasks);
    if (result.data.next_page_token) {
      return listAllTasks($http, tasks, result.data.next_page_token);
    } else {
      return tasks
    }
  });
}

app.controller('TaskListController', function($scope, NgTableParams, $http) {
  $scope.shortID = shortID;
  $scope.tasks = [];

  listAllTasks($http).then(function(tasks) {
    console.log(tasks);
    $scope.tasks = tasks;
    $scope.tableParams = new NgTableParams(
      {
        count: 25
      }, 
      {
        counts: [25, 50, 100],
        paginationMinBlocks: 2,
        paginationMaxBlocks: 10,
        total: tasks.length,
        dataset: tasks,
      }
    );
  });

  $scope.cancelTask = function(taskID) {
    var url = "/v1/tasks/" + taskID + ":cancel";
    $http.post(url);
  }
});

app.controller('WorkerListController', function($scope, $http) {

	$scope.url = "/v1/funnel/workers";
	$scope.workers = [];

  $http.get($scope.url).then(function(result) {
    var workers = result.data.workers || [];
console.log(workers)
    $scope.workers = workers;
  });
});

app.controller('TaskInfoController', function($scope, $http, $routeParams, $location) {
  
  $scope.url = "/v1/tasks/" + $routeParams.task_id

  $scope.task = {};
  $scope.cmdStr = function(cmd) {
    return cmd.join(' ');
  };

  $scope.fetchContent = function() {
    $http.get($scope.url).then(function(result){
      console.log(result.data);
      $scope.task = result.data
    })
  }
  $scope.fetchContent();

  $scope.serverURL = $location.protocol() + "://" + $location.host() + ":" + $location.port();

  $scope.cancelTask = function() {
    $http.post($scope.url + ":cancel");
  }
});

app.config(
  ['$routeProvider',
   function($routeProvider) {
     $routeProvider.when('/', {
       templateUrl: 'static/list.html',
     }).
     when('/tasks/:task_id', {
       templateUrl: 'static/task.html'
     }).
     when('/workers/', {
       templateUrl: 'static/worker-list.html'
     })
   }
  ]
);
