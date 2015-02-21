jQuery(document).ready(function() {
	jQuery("#streamcommands :input").click(function() {
		var postURL = $(this).attr('href');
		jQuery.ajax({	type:"POST",
									url:postURL
								});
	});

	jQuery("#chantoggle :input").click(function() {
		var postURL = $(this).attr('href') + 
									$(this).attr('value') + '/' + 
									$(this).is(':checked');
		jQuery.ajax({	type:"POST",
									url:postURL
								});
	});

	jQuery("#complexcommands #update").click(function() {
		var chan = $("form#longcommand #channel").val()
		var postURL = $(this).attr('href') + chan + "/x" + chan +
									$("form#longcommand #power").val() +
									$("form#longcommand #gain").val() +
									$("form#longcommand #input").val() +
									$("form#longcommand #bias").val() +
									$("form#longcommand #srb2").val() +
									$("form#longcommand #srb1").val() +
									"X"
		jQuery.ajax({	type:"POST",
									url:postURL
								});
	});

});
