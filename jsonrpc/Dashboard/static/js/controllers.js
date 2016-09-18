/**
 * INSPINIA - Responsive Admin Theme
 *
 * Main controller.js file
 * Define controllers with data used in Inspinia theme
 *
 *
 * Functions (controllers)
 *  - MainCtrl
 *  - dashboardFlotOne
 *  - dashboardFlotTwo
 *  - dashboardFlotFive
 *  - dashboardMap
 *  - flotChartCtrl
 *  - rickshawChartCtrl
 *  - sparklineChartCtrl
 *  - widgetFlotChart
 *  - modalDemoCtrl
 *  - ionSlider
 *  - wizardCtrl
 *  - CalendarCtrl
 *  - chartJsCtrl
 *  - GoogleMaps
 *  - ngGridCtrl
 *  - codeEditorCtrl
 *  - nestableCtrl
 *  - notifyCtrl
 *  - translateCtrl
 *  - imageCrop
 *  - diff
 *  - idleTimer
 *  - liveFavicon
 *  - formValidation
 *  - agileBoard
 *  - draggablePanels
 *  - chartistCtrl
 *  - metricsCtrl
 *  - sweetAlertCtrl
 *  - selectCtrl
 *  - toastrCtrl
 *  - loadingCtrl
 *  - datatablesCtrl
 *  - truncateCtrl
 *  - touchspinCtrl
 *  - tourCtrl
 *  - jstreeCtrl
 *
 *
 */

/**
 * MainCtrl - controller
 * Contains several global data used in different view
 *
 */
function MainCtrl() {

    /**
     * daterange - Used as initial model for data range picker in Advanced form view
     */
    this.daterange = {startDate: null, endDate: null};

    /**
     * slideInterval - Interval for bootstrap Carousel, in milliseconds:
     */
    this.slideInterval = 5000;


    /**
     * states - Data used in Advanced Form view for Chosen plugin
     */
    this.states = [
        'Alabama',
        'Alaska',
        'Arizona',
        'Arkansas',
        'California',
        'Colorado',
        'Connecticut',
        'Delaware',
        'Florida',
        'Georgia',
        'Hawaii',
        'Idaho',
        'Illinois',
        'Indiana',
        'Iowa',
        'Kansas',
        'Kentucky',
        'Louisiana',
        'Maine',
        'Maryland',
        'Massachusetts',
        'Michigan',
        'Minnesota',
        'Mississippi',
        'Missouri',
        'Montana',
        'Nebraska',
        'Nevada',
        'New Hampshire',
        'New Jersey',
        'New Mexico',
        'New York',
        'North Carolina',
        'North Dakota',
        'Ohio',
        'Oklahoma',
        'Oregon',
        'Pennsylvania',
        'Rhode Island',
        'South Carolina',
        'South Dakota',
        'Tennessee',
        'Texas',
        'Utah',
        'Vermont',
        'Virginia',
        'Washington',
        'West Virginia',
        'Wisconsin',
        'Wyoming'
    ];

    /**
     * check's - Few variables for checkbox input used in iCheck plugin. Only for demo purpose
     */
    this.checkOne = true;
    this.checkTwo = true;
    this.checkThree = true;
    this.checkFour = true;

    /**
     * knobs - Few variables for knob plugin used in Advanced Plugins view
     */
    this.knobOne = 75;
    this.knobTwo = 25;
    this.knobThree = 50;

    /**
     * Variables used for Ui Elements view
     */
    this.bigTotalItems = 175;
    this.bigCurrentPage = 1;
    this.maxSize = 5;
    this.singleModel = false;
    this.radioModel = 'Middle';
    this.checkModel = {
        left: false,
        middle: true,
        right: false
    };

    /**
     * groups - used for Collapse panels in Tabs and Panels view
     */
    this.groups = [
        {
            title: 'Dynamic Group Header - 1',
            content: 'Dynamic Group Body - 1'
        },
        {
            title: 'Dynamic Group Header - 2',
            content: 'Dynamic Group Body - 2'
        }
    ];

    /**
     * alerts - used for dynamic alerts in Notifications and Tooltips view
     */
    this.alerts = [
        { type: 'danger', msg: 'Oh snap! Change a few things up and try submitting again.' },
        { type: 'success', msg: 'Well done! You successfully read this important alert message.' },
        { type: 'info', msg: 'OK, You are done a great job man.' }
    ];

    /**
     * addAlert, closeAlert  - used to manage alerts in Notifications and Tooltips view
     */
    this.addAlert = function() {
        this.alerts.push({msg: 'Another alert!'});
    };

    this.closeAlert = function(index) {
        this.alerts.splice(index, 1);
    };

    /**
     * randomStacked - used for progress bar (stacked type) in Badges adn Labels view
     */
    this.randomStacked = function() {
        this.stacked = [];
        var types = ['success', 'info', 'warning', 'danger'];

        for (var i = 0, n = Math.floor((Math.random() * 4) + 1); i < n; i++) {
            var index = Math.floor((Math.random() * 4));
            this.stacked.push({
                value: Math.floor((Math.random() * 30) + 1),
                type: types[index]
            });
        }
    };
    /**
     * initial run for random stacked value
     */
    this.randomStacked();

    /**
     * summernoteText - used for Summernote plugin
     */
    this.summernoteText = ['<h3>Hello Jonathan! </h3>',
    '<p>dummy text of the printing and typesetting industry. <strong>Lorem Ipsum has been the dustrys</strong> standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more',
        'recently with</p>'].join('');

    /**
     * General variables for Peity Charts
     * used in many view so this is in Main controller
     */
    this.BarChart = {
        data: [5, 3, 9, 6, 5, 9, 7, 3, 5, 2, 4, 7, 3, 2, 7, 9, 6, 4, 5, 7, 3, 2, 1, 0, 9, 5, 6, 8, 3, 2, 1],
        options: {
            fill: ["#1ab394", "#d7d7d7"],
            width: 100
        }
    };

    this.BarChart2 = {
        data: [5, 3, 9, 6, 5, 9, 7, 3, 5, 2],
        options: {
            fill: ["#1ab394", "#d7d7d7"]
        }
    };

    this.BarChart3 = {
        data: [5, 3, 2, -1, -3, -2, 2, 3, 5, 2],
        options: {
            fill: ["#1ab394", "#d7d7d7"]
        }
    };

    this.LineChart = {
        data: [5, 9, 7, 3, 5, 2, 5, 3, 9, 6, 5, 9, 4, 7, 3, 2, 9, 8, 7, 4, 5, 1, 2, 9, 5, 4, 7],
        options: {
            fill: '#1ab394',
            stroke: '#169c81',
            width: 64
        }
    };

    this.LineChart2 = {
        data: [3, 2, 9, 8, 47, 4, 5, 1, 2, 9, 5, 4, 7],
        options: {
            fill: '#1ab394',
            stroke: '#169c81',
            width: 64
        }
    };

    this.LineChart3 = {
        data: [5, 3, 2, -1, -3, -2, 2, 3, 5, 2],
        options: {
            fill: '#1ab394',
            stroke: '#169c81',
            width: 64
        }
    };

    this.LineChart4 = {
        data: [5, 3, 9, 6, 5, 9, 7, 3, 5, 2],
        options: {
            fill: '#1ab394',
            stroke: '#169c81',
            width: 64
        }
    };

    this.PieChart = {
        data: [1, 5],
        options: {
            fill: ["#1ab394", "#d7d7d7"]
        }
    };

    this.PieChart2 = {
        data: [226, 360],
        options: {
            fill: ["#1ab394", "#d7d7d7"]
        }
    };
    this.PieChart3 = {
        data: [0.52, 1.561],
        options: {
            fill: ["#1ab394", "#d7d7d7"]
        }
    };
    this.PieChart4 = {
        data: [1, 4],
        options: {
            fill: ["#1ab394", "#d7d7d7"]
        }
    };
    this.PieChart5 = {
        data: [226, 134],
        options: {
            fill: ["#1ab394", "#d7d7d7"]
        }
    };
    this.PieChart6 = {
        data: [0.52, 1.041],
        options: {
            fill: ["#1ab394", "#d7d7d7"]
        }
    };
};


/**
 * translateCtrl - Controller for translate
 */
function translateCtrl($translate, $scope) {
    $scope.changeLanguage = function (langKey) {
        $translate.use(langKey);
        $scope.language = langKey;
    };
}


/**
 * diff - Controller for diff function
 */
function diff($scope) {
    $scope.oldText = 'Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry\'s standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only centuries, but also the leap into electronic typesetting.';
    $scope.newText = 'Lorem Ipsum is simply typesetting dummy text of the printing and has been the industry\'s typesetting. Lorem Ipsum has been the industry\'s standard dummy text ever the 1500s, when an printer took a galley of type and simply it to make a type. It has survived not only five centuries, but survived not also the leap into electronic typesetting.';

    $scope.oldText1 = 'Lorem Ipsum is simply printing and typesetting industry. Lorem Ipsum has been the industry\'s standard dummy text eve';
    $scope.newText1 = 'Ting dummy text of the printing and has been the industry\'s typesetting. Lorem Ipsum has been the industry\'s';
}

function datatables($scope,DTOptionsBuilder){

    $scope.dtOptions = DTOptionsBuilder.newOptions()
        .withOption('order', [0, 'desc'])
        .withDOM('<"html5buttons"B>lTfgitp')
        .withButtons([
            // {extend: 'copy'},
            // {extend: 'csv'},
            // {extend: 'excel', title: 'ExampleFile'},
            // {extend: 'pdf', title: 'ExampleFile'},
            //
            // {extend: 'print',
            //     customize: function (win){
            //         $(win.document.body).addClass('white-bg');
            //         $(win.document.body).css('font-size', '10px');
            //
            //         $(win.document.body).find('table')
            //             .addClass('compact')
            //             .css('font-size', 'inherit');
            //     }
            // }
        ]);
}

function SummaryCtrl($scope, $rootScope, SummaryService) {

    SummaryService.getLastestBlock()
        .then(function(res){
            $scope.number = res.number;
            $rootScope.height = res.number;

            SummaryService.getAvgTimeAndCount("1",res.number+"")
                .then(function(res){
                    if (res.time < 0) {
                        $scope.avgTime = 0
                    } else {
                        $scope.avgTime = res.time;
                    }
                    // $scope.txCount = res.count; // 后端没有存到数据库里
                    $scope.txCount = $scope.number * 500;
                }, function(error){
                    console.log(error);
                })

        }, function(error){
            console.log(error)
        })

    SummaryService.getNodeInfo()
        .then(function(res){
            $scope.nodes = res;
        }, function(error){
            console.log(error)
        })
}

function BlockCtrl($scope, DTOptionsBuilder, SummaryService, BlockService, TransactionService) {
    $scope.status = "";

    $scope.tx = {
        from: "6201cb0448964ac597faf6fdf1f472edf2a22b89",
        to: "0000000000000000000000000000000000000002",
        value: "1"
    };

    $scope.blockAvg = {
        from: "",
        to: ""
    };

    $scope.block = {
        from:"",
        to:""
    };

    $scope.blockEvm = {
        from: "",
        to: ""
    };

    $scope.commitTime = "0";
    $scope.batchTime = "0";
    $scope.avgTime = "0";
    $scope.evmTime = "0";

    var getBlocks = function() {
        BlockService.getAllBlocks()
            .then(function(res){
                $scope.blocks = res;
            }, function(error){
                console.log(error);
            })
    };

    datatables($scope, DTOptionsBuilder);
    getBlocks();


    $scope.submit = function(){

        if (isEmpty($scope.tx)) {
            alert("字段不能为空");
            return false;
        }

        $scope.status = "please waitting.....";
        TransactionService.SendTransaction($scope.tx.from, $scope.tx.to, $scope.tx.value)
            .then(function(res){
                $scope.status = res;
                getBlocks();
            }, function(error){
                $scope.status = error.message;
                console.log(error);
            })
    };

    $scope.queryAvg = function(){

        if (isEmpty($scope.blockAvg)) {
            alert("字段不能为空");
            return false;
        }

        SummaryService.getAvgTimeAndCount($scope.blockAvg.from, $scope.blockAvg.to)
            .then(function(res){
                $scope.avgTime = res.time
            }, function(error){
                console.log(error);
            })
    }

    $scope.query = function(){

        if (isEmpty($scope.block)) {
            alert("字段不能为空");
            return false;
        }

        BlockService.queryCommitAndBatchTime($scope.block.from, $scope.block.to)
            .then(function(res){
                $scope.commitTime = res.CommitTime;
                $scope.batchTime = res.BatchTime
            }, function(error){
                // $scope.status = "error";
                console.log(error);
            })
    }

    $scope.queryEvmTime = function() {

        if (isEmpty($scope.blockEvm)) {
            alert("字段不能为空");
            return false;
        }

        BlockService.queryEvmAvgTime($scope.blockEvm.from, $scope.blockEvm.to)
            .then(function(res){
                $scope.evmTime = res;
            }, function(error){
                console.log(error);
            })
    }
}

function TransactionCtrl($scope, DTOptionsBuilder, TransactionService) {
    $scope.status = "";

    datatables($scope, DTOptionsBuilder);

    TransactionService.getAllTxs()
        .then(function(res){
            $scope.txs = res;
        }, function(error){
            console.log(error);
        })
}

function AccountCtrl($scope, DTOptionsBuilder, AccountService) {

    $scope.account = {
        password: "",
        confirm:""
    };
    datatables($scope, DTOptionsBuilder);

    AccountService.getAllAccounts()
        .then(function(res){
            $scope.accounts = res;
        } ,function(error){
            console.log(error);
        })
    $scope.submit = function(){

        if (isEmpty($scope.account)) {
            alert("字段不能为空");
            return false;
        }
        if($scope.account.password != $scope.account.confirm) {
            alert("两次输入密码不一致！")
            $scope.account.password = "";
            $scope.account.confirm = "";
            return false;
        }
        // $scope.status = "please waitting.....";
        AccountService.newAccount($scope.account.password)
            .then(function(res){
                $scope.address = res;
            }, function(error){
                // $scope.status = error.message;
                console.log(error);
            })
    };
}

function AddProjectCtrl($scope, $state, $cookies, ENV, ContractService) {

    $scope.flag = false;

    $scope.PATTERN = ENV.PATTERN;
    
    $scope.project = {
        name: "",
        type: "1",
        pattern: "",
        abi: []
    };

    $scope.disable = false;

    $scope.select = function(){
        $scope.disable = false;
        $scope.flag = false;
        $scope.project.abi = [];
    }

    $scope.compile = function(){
        console.log($scope.project)
        if (isEmpty($scope.project)) {
            alert("字段不能为空");
            return false;
        }

        $scope.disable = true;
        ContractService.compileContract($scope.project.pattern.value)
            .then(function(res){
                $scope.flag = true;
                var abis = [];

                for (var i = 0;i < res.length; i++) {
                    abis.push(JSON.parse(res[i]))
                }

                $scope.project.abi = abis

            }, function(error){
                alert(error.message);
                console.log(error);
            })
    }

    $scope.saveABI = function() {
            console.log($scope.project);

            // todo 现有合约个数
            var cookieValue = $cookies.getObject(ENV.COOKIE);
            var len;
            // var len = ENV.CONTRACT.length;

            // contract
            for (var i = 0;i < $scope.project.abi.length; i++) {

                var _contract = {};

                _contract.projectName = $scope.project.name;
                _contract.type = $scope.project.type;    // 1: Create 2: Load
                // _contract.contractName = $scope.project.abi[i].  // 如何得到合约名字？？
                // len++;
                // _contract.contractName = "Contract_"+ len;  // 如何得到合约名字？？正则？
                // _contract.methods = [];
                _contract.methods = $scope.project.abi[i];
                _contract.status = 0; // 0: Nondeployed 1: Deployed
                _contract.sourceCode = $scope.project.pattern.value;
                _contract.hash = "";

                // contract["Contract_"+ len] = _contract;
                // console.log(contract);
                // ENV.CONTRACT.push(contract)
                // todo 将合约存到cookie或文件中
                console.log(cookieValue)
                if (!cookieValue) {
                    len = 1;
                    _contract.contractName = "Contract_"+ len;  // 如何得到合约名字？？正则？
                    var objContract = _defineProperty({}, _contract.contractName, _contract);
                    $cookies.putObject("contracts",objContract)
                } else {
                    len = Object.keys(cookieValue).length;
                    len++;
                    _contract.contractName = "Contract_"+ len;  // 如何得到合约名字？？正则？
                    cookieValue[_contract.contractName] = _contract;
                    $cookies.putObject(ENV.COOKIE, cookieValue);
                }

                // ENV.CONTRACT.push(_contract)
            }

        $state.go("dashboards.contract")
    }
}


function ContractCtrl($scope, $uibModal, $cookies, DTOptionsBuilder, SweetAlert, ENV) {

    // todo 从cookie或文件中取出所有合约
    $scope.contracts = $cookies.getObject(ENV.COOKIE);

    // $scope.contracts = ENV.CONTRACT;
    $scope.contract = {
        from: ENV.FROM
    };

    datatables($scope, DTOptionsBuilder);

    $scope.modal_deploy = function (ctName, sourceCode) {
        $scope.ctName = ctName;
        $scope.sourceCode = sourceCode;
        var modalInstance = $uibModal.open({
            templateUrl: 'static/views/modal_deploy.html',
            controller: modalInstanceCtrl,
            scope: $scope
        });
    };

    $scope.modal_invoke = function(ctHash, methods) {
        $scope.ctHash = ctHash;
        $scope.methods = methods;
        var modalInstance = $uibModal.open({
            templateUrl: 'static/views/modal_invoke.html',
            controller: modalInstanceInvodeCtrl,
            scope: $scope
        });
    };

    $scope.delete = function(name){
        SweetAlert.swal({
                title: "Are you sure?",
                text: "Your will delete the contract from cookie!",
                type: "warning",
                showCancelButton: true,
                confirmButtonColor: "#DD6B55",
                confirmButtonText: "Yes, delete it!",
                closeOnConfirm: false,
                closeOnCancel: false
            },
            function (isConfirm) {
                if (isConfirm) {
                    var cookieValue = $cookies.getObject(ENV.COOKIE);
                    delete cookieValue[name]
                    delete $scope.contracts[name]
                    $cookies.putObject(ENV.COOKIE, cookieValue)
                    SweetAlert.swal("Deleted!", "The contract has deleted from cookie.", "success");
                } else {
                    SweetAlert.swal("Cancelled", ":)", "error");
                }
            });
    }
}

function modalInstanceCtrl ($scope, $uibModalInstance, $cookies, SweetAlert, ENV, ContractService) {

    $scope.ok = function () {
        ContractService.deployContract($scope.from,$scope.sourceCode)
            .then(function(res){
                var cookieValue = $cookies.getObject(ENV.COOKIE);
                for (var name in cookieValue) {
                    if ( name == $scope.ctName) {
                        cookieValue[name].status = 1;
                        cookieValue[name].hash = res;

                        $scope.contracts[name] = cookieValue[name];
                        $cookies.putObject(ENV.COOKIE, cookieValue)

                        break;
                    }
                }
                // for (var i = 0;i < ENV.CONTRACT.length; i++) {
                //         if ( ENV.CONTRACT[i].contractName == $scope.ctName) {
                //             ENV.CONTRACT[i].status = 1;
                //             ENV.CONTRACT[i].hash = res;
                //             console.log(ENV.CONTRACT[i])
                //             break;
                //         }
                // }
                SweetAlert.swal({
                    title: "Deployed successfully!",
                    text: "The contract hash is <span class='text_red'>"+res+"</span>",
                    type: "success",
                    customClass: 'swal-wide',
                    html: true
                });
            }, function(err){
                console.log(err)
            });

        SweetAlert.swal("Deployed!", "You have deployed the contract successfully!", "success");
        $uibModalInstance.close();
    };

    $scope.cancel = function () {
        SweetAlert.swal("Cancelled", "You don't deploy the contract :)", "error");
        $uibModalInstance.dismiss('cancel');
    };

}


function modalInstanceInvodeCtrl ($scope, $uibModalInstance, SweetAlert, ENV, ContractService, UtilsService) {
    console.log($scope.methods);
    var abimethod = {};

    $scope.method = {
        name: $scope.methods[0].name,
        params: {}
    };

    $scope.submit = function () {

        for (var i = 0;i < $scope.methods.length;i++) {
            if ($scope.methods[i].name === $scope.method.name) {
                abimethod = $scope.methods[i];
                break;
            }
        }

        UtilsService.encode(abimethod,$scope.method.params)
            .then(function(res) {
                // 调用合约
                console.log(res);

                // from 调用者地址，to 合约地址，data 为编码
                ContractService.invokeContract(ENV.FROM,  $scope.ctHash, res)
                    .then(function(res){
                        // $scope.status = res;
                        // getBlocks();

                        SweetAlert.swal({
                            title: "Invoked successfully!",
                            text: "You have invode the <span class='text_red'>"+ $scope.method.name +"</span> method of contract successfully! The address is <span class='text_red'>"+ res +"</span>",
                            type: "success",
                            customClass: 'swal-wide',
                            html: true
                        });
                        $uibModalInstance.close();
                    }, function(error){
                        // $scope.status = error.message;
                        console.log(error);
                        SweetAlert.swal("Error！", "", "error");
                        $uibModalInstance.close();
                    })
            }, function(err) {
                console.log(err);
                SweetAlert.swal("Error！", "", "error");
                $uibModalInstance.close();
            });
    };

    $scope.cancel = function () {
        SweetAlert.swal("Cancelled", "You don't invoke the contract :)", "error");
        $uibModalInstance.dismiss('cancel');
    };
}

function isEmpty(obj) {
    for (var key in obj) {
        if (!obj[key]) {
            return true
        }
    }
    return false
}

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

/**
 *
 * Pass all functions into module
 */
angular
    .module('starter')
    .controller('MainCtrl', MainCtrl)
    .controller('translateCtrl', translateCtrl)
    .controller('SummaryCtrl', SummaryCtrl)
    .controller('BlockCtrl', BlockCtrl)
    .controller('TransactionCtrl',TransactionCtrl)
    .controller('AccountCtrl', AccountCtrl)
    .controller('ContractCtrl', ContractCtrl)
    .controller('AddProjectCtrl', AddProjectCtrl)

