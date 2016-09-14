/**
 * Created by sammy on 16-8-16.
 */

var selectedClassName = "selected";
var flag = true;

$(document).ready(function(){

    var $tab_tx    = $("#tab_tx"),
        $tab_bal   = $("#tab_bal"),
        $tab_block = $("#tab_block");

    var $div_trans    =  $("div.trans"),
        $div_balances = $("div.balances"),
        $div_blocks   = $("div.blocks");

    var submitForm = function(){
        var $form = $("form");
        var data = getFormData($form);
        var count = data.count;

        if (!count) {
            count = 1
        }

        $(".status").html("请求发送中...")

        for(var i = 1;i <= count;i++) {

            $.ajax({
                contentType: "application/json",
                type: "POST",
                dataType: "json",
                url: "/trans",
                data: JSON.stringify(data),
                success: function (result) {
                    console.log(result);
                    if (result.Code == 1) {
                        //alert("提交成功!");
                        //alert("提交成功!");
                        $(".status").html("提交成功")
                    } else {
                        $(".status").html("提交失败，余额不足");
                    }

                    // location.reload();
                },
                error: function (err) {
                    if (err) {
                        console.log(err);
                        $("input[type=reset]").trigger("click");
                    }
                    return false;
                }
            });
        }
    };

    $.extend($.fn.pagination.defaults, {
        pageSize: 50,
        className: 'paginationjs-big'
    });

    $("#submit").click(submitForm);

    $("#query").click(function(){
        $.ajax({
            contentType: "application/json",
            type: "POST",
            dataType: "json",
            url: "/query",
            data: JSON.stringify({
                from: $("input[name='from_block']").val(),
                to: $("input[name='to_block']").val()
            }),
            success: function (result) {
                if (result.Code == 1) {
                    $("#time").html(result.Data)
                } 
            },
            error: function (err) {
                if (err) {
                    console.log(err);
                }
                return false;
            }
        });
    });

    $("#commitandbatchquery").click(function(){
        $.ajax({
            contentType: "application/json",
            type: "POST",
            dataType: "json",
            url: "/commitandbatchquery",
            data: JSON.stringify({
                from: $("input[name='from_block2']").val(),
                to: $("input[name='to_block2']").val()
            }),
            success: function (result) {
                if (result.Code == 1) {
                    $("#committime").html(result.Commit)
                    $("#batchtime").html(result.Batch)
                }
            },
            error: function (err) {
                if (err) {
                    console.log(err);
                }
                return false;
            }
        });
    });

    $tab_tx.on('click',{div: $div_trans, url: "/trans"},tabClickHandler).trigger('click');
    $tab_bal.on('click',{div: $div_balances, url: "/balances"},tabClickHandler);
    $tab_block.on('click',{div: $div_blocks, url: "/blocks"},tabClickHandler);



    // $tab_tx.click({div: $div_trans, url: "/trans"},tabClickHandler);
    // $tab_bal.click({div: $div_balances, url: "/balances"},tabClickHandler);
    // $tab_block.click({div: $div_blocks, url: "/blocks"},tabClickHandler);

});


function getFormData($form){
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}

// Tab click event Handler
function tabClickHandler(event) {

    if (flag) {
        console.log(event.data);
        var _this = $(this);

        var $div = $(event.data.div[0]);
        var $tbody = $div.find("tbody");

        flag = false;

        _this.addClass(selectedClassName).siblings().removeClass(selectedClassName);
        $div.show().siblings().hide();
        $tbody.html("数据加载中，请稍等......");

        setTimeout(function(){
            $.ajax({
                contentType: "application/json",
                type: "GET",
                dataType: "json",
                url: event.data.url,
                success: function (result) {
                    // console.log(result);
                    flag = true;
                    if (result.Code == 0) {
                        alert("error");
                        return
                    }
                    // var $tbody = $div.find("tbody");
                    $tbody.html("");

                    if (result.Data instanceof Array) {
                        // Array
                        $div.find('.pagination-container').pagination({
                            dataSource: result.Data,
                            // pageSize: 50,
                            callback: function(data, pagination) {
                                var html = render(data);
                                $div.find('.data-container').html(html);
                            }
                        });
                    } else {
                        // Object
                        for (var key in result.Data) {
                            var $tr = $("<tr>");
                            var $td1 = $("<td>");
                            var $td2 = $("<td>");

                            $td1.html(key);
                            $td2.html(result.Data[key]);

                            $tr.append($td1);
                            $tr.append($td2);

                            $tbody.append($tr);
                        }
                    }
                },
                error: function (err) {
                    if (err) {
                        console.log(err);
                    }
                    flag = true;
                    return false;
                }
            });
        },500);

    }
}

function render(data){
    var html = '';

    for (var i = 0; i < data.length; i++) {
        html += "<tr>";

        for (var col in data[i]) {
            html += "<td>"+data[i][col]+"</td>";
        }

        html += "</tr>"
    }
    return html;
}