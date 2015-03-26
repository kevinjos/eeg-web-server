jQuery(document).ready(function() {
	jQuery("#streamcommands :input").click(function() {
		var postURL = $(this).attr('href');
		jQuery.ajax({	type:"POST",
									url:postURL
								});
	});

	jQuery(".thumbnail").click(function() {
		$(this).toggleClass('faded');

		if ($(this).hasClass('faded')) {
			$(this).attr('active', false);
		} else {
			$(this).attr('active', true);
		};

		var button = $(this).find(":input")

		var postURL = $(button).attr('href') + 
									$(button).attr('value') + '/' + 
									$(this).attr('active');
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

	jQuery("#deltafft #deltafftupdate").click(function() {
		var fftsize = $("form#deltafftints #fftsize").val()
		//window.updatenewsize(fftsize)
		var fftfreq = $("form#deltafftints #fftfreq").val()
    //considering tracking state of delta bins via fftfreq
		var postURL = $(this).attr('href') + fftsize + "&" + fftfreq
		jQuery.ajax({	type:"POST",
									url:postURL
								});
	});
});
