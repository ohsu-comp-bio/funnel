"use strict";
var app = angular.module('TESApp', ['ngRoute'])

function shortID(longID) {
  return longID.split('-')[0];
}

app.controller('JobListController', function($scope, $http) {

		$scope.jobs = [];
    $scope.shortID = shortID;

		$scope.fetchContent = function() {
			$http.get("/v1/jobs").then(function(result){
				$scope.jobs = result.data.jobs;
			});
		}

		$scope.fetchContent();
});

app.controller('WorkerListController', function($scope, $http) {

		$scope.url = "/v1/jobs-service";
		$scope.workers = [];

		$scope.fetchContent = function() {
			$http.get($scope.url).then(function(result){
				$scope.workers = result.data;
			});
		}

		$scope.fetchContent();
});

app.controller('JobInfoController',
    function($scope, $http, $routeParams) {
        $scope.url = "/v1/jobs/" + $routeParams.job_id

        $scope.job = {};
        $scope.cmdStr = function(cmd) {
            return cmd.join(' ');
        };
        $scope.fetchContent = function() {
            $http.get($scope.url).then(function(result){
            console.log(result.data);
                $scope.job = result.data
            })
        }
        $scope.fetchContent();
    }
);

app.config(['$routeProvider',
    function($routeProvider) {
        $routeProvider.when('/', {
           templateUrl: 'static/list.html',
        }).
        when('/jobs/:job_id', {
           templateUrl: 'static/job.html'
        })
    }
]);
