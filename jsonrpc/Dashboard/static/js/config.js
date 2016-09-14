/**
 * INSPINIA - Responsive Admin Theme
 *
 * Inspinia theme use AngularUI Router to manage routing and views
 * Each view are defined as state.
 * Initial there are written state for all view in theme.
 *
 */
function config($stateProvider, $urlRouterProvider, $ocLazyLoadProvider, IdleProvider) {

    // Configure Idle settings
    IdleProvider.idle(5); // in seconds
    IdleProvider.timeout(120); // in seconds

    $urlRouterProvider.otherwise("/dashboards/summary");

    $ocLazyLoadProvider.config({
        // Set to true if you want to see what and when is dynamically loaded
        debug: false
    });

    $stateProvider

        .state('dashboards', {
            abstract: true,
            url: "/dashboards",
            templateUrl: "static/views/common/content.html"
        })

        .state('dashboards.summary', {
            url: "/summary",
            templateUrl: "static/views/summary.html",
            data: { pageTitle: 'summary' }
        })
        .state('dashboards.contract', {
            url: "/contract",
            templateUrl: "static/views/contract.html",
            data: { pageTitle: 'Contract' },
            resolve: {
                loadPlugin: function ($ocLazyLoad) {
                    return $ocLazyLoad.load([
                        {
                            files: ['static/js/plugins/footable/footable.all.min.js', 'static/css/plugins/footable/footable.core.css']
                        },
                        {
                            name: 'ui.footable',
                            files: ['static/js/plugins/footable/angular-footable.js']
                        },
                        {
                            files: ['static/js/plugins/sweetalert/sweetalert.min.js', 'static/css/plugins/sweetalert/sweetalert.css']
                        },
                        {
                            name: 'oitozero.ngSweetAlert',
                            files: ['static/js/plugins/sweetalert/angular-sweetalert.min.js']
                        }
                    ]);
                }
            }
        })
        .state('contract', {
            abstract: true,
            url: "/contract",
            templateUrl: "static/views/common/content.html"
        })
        .state('contract.addProject', {
            url: "/add",
            templateUrl: "static/views/addProject.html",
            data: { pageTitle: 'Project' }
            ,
            resolve: {
                loadPlugin: function ($ocLazyLoad) {
                    return $ocLazyLoad.load([
                        {
                            name: 'ui.select',
                            files: ['static/js/plugins/ui-select/select.min.js', 'static/css/plugins/ui-select/select.min.css']
                        }
                    ]);
                }
            }
        })
        .state('contract.invokeContract', {
            url: "/invoke",
            templateUrl: "static/views/invokeContract.html",
            data: { pageTitle: 'Invoke' }
            ,
            resolve: {
                loadPlugin: function ($ocLazyLoad) {
                    return $ocLazyLoad.load([
                        {
                            name: 'ui.select',
                            files: ['static/js/plugins/ui-select/select.min.js', 'static/css/plugins/ui-select/select.min.css']
                        }
                    ]);
                }
            }
        })
        .state('blockchain', {
            abstract: true,
            url: "/blockchain",
            templateUrl: "static/views/common/content.html"
        })
        .state('blockchain.block_tables', {
            url: "/blocks",
            templateUrl: "static/views/block.html",
            data: { pageTitle: 'Block' },
            resolve: {
                loadPlugin: function ($ocLazyLoad) {
                    return $ocLazyLoad.load([
                        {
                            files: ['static/js/plugins/footable/footable.all.min.js', 'static/css/plugins/footable/footable.core.css']
                        },
                        {
                            name: 'ui.footable',
                            files: ['static/js/plugins/footable/angular-footable.js']
                        }
                    ]);
                }
            }
        })
        .state('blockchain.tx_tables', {
            url: "/transactions",
            templateUrl: "static/views/transaction.html",
            data: { pageTitle: 'Transaction' },
            resolve: {
                loadPlugin: function ($ocLazyLoad) {
                    return $ocLazyLoad.load([
                        {
                            files: ['static/js/plugins/footable/footable.all.min.js', 'static/css/plugins/footable/footable.core.css']
                        },
                        {
                            name: 'ui.footable',
                            files: ['static/js/plugins/footable/angular-footable.js']
                        }
                    ]);
                }
            }
        })
        .state('blockchain.account_tables', {
            url: "/accounts",
            templateUrl: "static/views/account.html",
            data: { pageTitle: 'Account' },
            resolve: {
                loadPlugin: function ($ocLazyLoad) {
                    return $ocLazyLoad.load([
                        {
                            files: ['static/js/plugins/footable/footable.all.min.js', 'static/css/plugins/footable/footable.core.css']
                        },
                        {
                            name: 'ui.footable',
                            files: ['static/js/plugins/footable/angular-footable.js']
                        }
                    ]);
                }
            }
        })
}
angular
    .module('inspinia')
    .config(config)
    .run(function($rootScope, $state) {
        $rootScope.$state = $state;
    });
